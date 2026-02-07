package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/baking-bad/noble-indexer/internal/cache"
	"github.com/dipdup-io/ipfs-tools"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	FileSizeLimit = 10 * 1024 * 1024
	SubstrIpfs    = "/ipfs/"
)

type ContractMetadata struct {
	Compiler Compiler          `json:"compiler"`
	Language string            `json:"language"`
	Output   Output            `json:"output"`
	Settings Settings          `json:"settings"`
	Sources  map[string]Source `json:"sources"`
}

type TokenMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Attributes  []any  `json:"attributes"`
}

type Compiler struct {
	Version string `json:"version"`
}

type Output struct {
	ABI     json.RawMessage `json:"abi"`
	Devdoc  Doc             `json:"devdoc"`
	Userdoc Doc             `json:"userdoc"`
}

type Doc struct {
	Kind    string                          `json:"kind"`
	Title   string                          `json:"title,omitempty"`
	Version int                             `json:"version"`
	Notice  string                          `json:"notice,omitempty"`
	Methods map[string]DocMethodDescription `json:"methods"`
}

type DocMethodDescription struct {
	Notice  string            `json:"notice,omitempty"`
	Details string            `json:"details,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	Returns map[string]string `json:"returns,omitempty"`
}

type Source struct {
	Keccak256 string   `json:"keccak256"`
	License   string   `json:"license,omitempty"`
	Urls      []string `json:"urls,omitempty"`
	Content   string   `json:"content,omitempty"`
}

type Settings struct {
	CompilationTarget map[string]string `json:"compilationTarget"`
	EvmVersion        string            `json:"evmVersion"`
	Libraries         map[string]string `json:"libraries"`
	Metadata          MetadataInfo      `json:"metadata"`
	Optimizer         OptimizerSettings `json:"optimizer"`
	Remappings        []string          `json:"remappings"`
}

type MetadataInfo struct {
	BytecodeHash string `json:"bytecodeHash"`
}

type OptimizerSettings struct {
	Enabled bool `json:"enabled"`
	Runs    int  `json:"runs"`
}

type Pool struct {
	ipfs  ipfs.IPool
	cache cache.ICache
}

func New(gateways string, opts ...Option) (Pool, error) {
	sources := strings.Split(gateways, ";")
	p, err := ipfs.NewPool(
		sources,
		FileSizeLimit,
	)
	if err != nil {
		return Pool{}, errors.Wrap(err, "creating ipfs pool")
	}
	pool := Pool{
		ipfs: p,
	}

	for _, opt := range opts {
		opt(&pool)
	}

	return pool, nil
}

func (p Pool) ValidateURL(link *url.URL) error {
	host := link.Host
	if strings.Contains(host, ":") {
		newHost, _, err := net.SplitHostPort(link.Host)
		if err != nil {
			return err
		}
		host = newHost
	}
	if host == "localhost" || host == "127.0.0.1" {
		return errors.Wrap(ErrInvalidURI, fmt.Sprintf("invalid host: %s", host))
	}

	for _, mask := range []string{
		"10.0.0.0/8",
		"100.64.0.0/10",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"240.0.0.0/4",
	} {
		_, cidr, err := net.ParseCIDR(mask)
		if err != nil {
			return err
		}

		ip := net.ParseIP(host)
		if ip != nil && cidr.Contains(ip) {
			return errors.Wrap(ErrInvalidURI, fmt.Sprintf("restricted subnet: %s", mask))
		}
	}
	return nil
}

func (p Pool) ContractMetadata(ctx context.Context, cid string) (ContractMetadata, error) {
	raw, err := p.LoadMetadata(ctx, cid)
	if err != nil {
		return ContractMetadata{}, err
	}

	var md ContractMetadata
	err = json.Unmarshal(raw, &md)
	return md, err
}

func (p Pool) TokenMetadata(ctx context.Context, cid string) (TokenMetadata, error) {
	raw, err := p.LoadMetadata(ctx, cid)
	if err != nil {
		return TokenMetadata{}, err
	}

	var md TokenMetadata
	err = json.Unmarshal(raw, &md)
	return md, err
}

func (p Pool) getFromCache(ctx context.Context, cid string) ([]byte, bool) {
	if p.cache == nil {
		return nil, false
	}
	val, ok := p.cache.Get(ctx, cid)
	if !ok {
		return nil, false
	}
	return []byte(val), true
}

func (p Pool) setToCache(ctx context.Context, cid string, data []byte) {
	if p.cache == nil {
		return
	}
	if err := p.cache.Set(ctx, cid, string(data), nil); err != nil {
		log.Err(err).Msg("setting to cache") // not critical, just log
	}
}

func (p Pool) LoadMetadata(ctx context.Context, cid string) ([]byte, error) {
	parsed, err := url.ParseRequestURI(cid)
	if err != nil {
		return nil, ErrInvalidURI
	}

	if err := p.ValidateURL(parsed); err != nil {
		return nil, err
	}

	if data, ok := p.getFromCache(ctx, cid); ok {
		return data, nil
	}

	path := ipfs.Path(cid)
	data, err := p.ipfs.Get(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "getting metadata")
	}

	if data.Raw == nil {
		return nil, errors.New("empty metadata")
	}

	p.setToCache(ctx, cid, data.Raw)
	return data.Raw, nil
}

func (p Pool) ContractText(ctx context.Context, urls []string) (contract string, err error) {
	for _, url := range urls {
		ipfsHash := extractIPFSHash(url)
		if ipfsHash == "" {
			continue
		}

		data, err := p.ipfs.Get(ctx, ipfsHash)
		if err != nil {
			return "", errors.Wrap(err, "getting contract text")
		}

		contract = string(data.Raw)
	}

	return
}

func extractIPFSHash(s string) string {
	idx := strings.Index(s, SubstrIpfs)
	if idx == -1 {
		return ""
	}
	return s[idx+len(SubstrIpfs):]
}

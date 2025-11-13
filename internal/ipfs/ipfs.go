package ipfs

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/dipdup-io/ipfs-tools"
	"github.com/pkg/errors"
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
	ipfs *ipfs.Pool
}

func New(gateways string) (Pool, error) {
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

	return pool, nil
}

func (p Pool) ContractMetadata(ctx context.Context, cid string) (metadata ContractMetadata, err error) {
	s := ipfs.Path(cid)
	data, err := p.ipfs.Get(ctx, s)
	if err != nil {
		return ContractMetadata{}, errors.Wrap(err, "getting metadata")
	}
	if data.Raw != nil {
		if err := json.Unmarshal(data.Raw, &metadata); err != nil {
			return ContractMetadata{}, errors.Wrap(err, "getting metadata")
		}
	}

	return metadata, nil
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

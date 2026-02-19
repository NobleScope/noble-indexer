package config

import (
	"github.com/NobleScope/noble-indexer/internal/profiler"
	"github.com/dipdup-net/go-lib/config"
	"github.com/pkg/errors"
)

type Config struct {
	config.Config            `yaml:",inline"`
	LogLevel                 string           `validate:"omitempty,oneof=debug trace info warn error fatal panic" yaml:"log_level"`
	Indexer                  Indexer          `yaml:"indexer"`
	API                      API              `yaml:"api"`
	Cache                    Cache            `yaml:"cache"`
	Profiler                 *profiler.Config `validate:"omitempty"                                               yaml:"profiler"`
	ContractMetadataResolver MetadataResolver `yaml:"contract_resolver"`
	TokenMetadataResolver    MetadataResolver `yaml:"token_resolver"`
	ContractVerifier         ContractVerifier `yaml:"contract_verifier"`
	Network                  string           `validate:"required"                                                yaml:"network"`
	Networks                 NetworksConfig   `yaml:"networks"`
}

type Indexer struct {
	Name            string         `validate:"omitempty"     yaml:"name"`
	StartLevel      int64          `validate:"omitempty"     yaml:"start_level"`
	BlockPeriod     int64          `validate:"omitempty"     yaml:"block_period"`
	ScriptsDir      string         `validate:"omitempty,dir" yaml:"scripts_dir"`
	AssetsDir       string         `validate:"omitempty,dir" yaml:"assets_dir"`
	GenesisFilename string         `validate:"omitempty"     yaml:"genesis_filename"`
	RequestBulkSize int            `validate:"min=1"         yaml:"request_bulk_size"`
	Proxy           ProxyContracts `yaml:"proxy_contracts"`
}

type API struct {
	Bind           string `validate:"required"        yaml:"bind"`
	RateLimit      int    `validate:"omitempty,min=0" yaml:"rate_limit"`
	RequestTimeout int    `validate:"omitempty,min=1" yaml:"request_timeout"`
	Websocket      bool   `validate:"omitempty"       yaml:"websocket"`
}

type Cache struct {
	URL string `validate:"required,url" yaml:"url"`
	TTL int    `validate:"min=1"        yaml:"ttl"`
}

type MetadataResolver struct {
	Name             string `validate:"omitempty" yaml:"name"`
	SyncPeriod       int64  `validate:"min=1"     yaml:"sync_period"`
	MetadataGateways string `validate:"required"  yaml:"metadata_gateways"`
	RequestBulkSize  int    `validate:"min=1"     yaml:"request_bulk_size"`
	RetryDelay       int    `validate:"min=1"     yaml:"retry_delay"`
	RetryCount       uint64 `validate:"omitempty" yaml:"retry_count"`
}

type ProxyContracts struct {
	Threads              int  `validate:"min=1" yaml:"threads"`
	SyncPeriodSeconds    int  `validate:"min=1" yaml:"sync_period_seconds"`
	BatchSize            int  `validate:"min=1" yaml:"node_batch_size"`
	MaxResolvingAttempts uint `validate:"min=1" yaml:"max_resolving_attempts"`
}

type ContractVerifier struct {
	SyncPeriod int64 `validate:"min=1" yaml:"sync_period"`
}

// Substitute -
func (c *Config) Substitute() error {
	if err := c.Config.Substitute(); err != nil {
		return err
	}
	return nil
}

type Network struct {
	PrecompiledContracts []string `validate:"omitempty,dive,eth_addr" yaml:"precompiled_contracts,omitempty"`
}

type NetworksConfig map[string]Network

func (nc NetworksConfig) Get(network string) (Network, error) {
	if netCfg, ok := nc[network]; ok {
		return netCfg, nil
	}
	return Network{}, errors.Errorf("network %s config not found", network)
}

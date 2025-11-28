package config

import (
	"github.com/baking-bad/noble-indexer/internal/profiler"
	"github.com/dipdup-net/go-lib/config"
)

type Config struct {
	config.Config            `yaml:",inline"`
	LogLevel                 string           `validate:"omitempty,oneof=debug trace info warn error fatal panic" yaml:"log_level"`
	Indexer                  Indexer          `yaml:"indexer"`
	API                      API              `yaml:"api"`
	Profiler                 *profiler.Config `validate:"omitempty"                                               yaml:"profiler"`
	ContractMetadataResolver MetadataResolver `yaml:"contract_resolver"`
	TokenMetadataResolver    MetadataResolver `yaml:"token_resolver"`
}

type Indexer struct {
	Name            string `validate:"omitempty"     yaml:"name"`
	StartLevel      int64  `validate:"omitempty"     yaml:"start_level"`
	BlockPeriod     int64  `validate:"omitempty"     yaml:"block_period"`
	ScriptsDir      string `validate:"omitempty,dir" yaml:"scripts_dir"`
	AssetsDir       string `validate:"omitempty,dir" yaml:"assets_dir"`
	RequestBulkSize int    `validate:"min=1"         yaml:"request_bulk_size"`
}

type API struct {
	Bind           string `validate:"required" yaml:"bind"`
	RateLimit      int    `validate:"min=0"    yaml:"rate_limit"`
	RequestTimeout int    `validate:"min=1"    yaml:"request_timeout"`
}

type MetadataResolver struct {
	Name             string `validate:"omitempty" yaml:"name"`
	SyncPeriod       int64  `validate:"min=1"     yaml:"sync_period"`
	MetadataGateways string `validate:"required"  yaml:"metadata_gateways"`
	RequestBulkSize  int    `validate:"min=1"     yaml:"request_bulk_size"`
	RetryDelay       int    `validate:"min=1"     yaml:"retry_delay"`
	RetryCount       uint64 `validate:"omitempty" yaml:"retry_count"`
}

// Substitute -
func (c *Config) Substitute() error {
	if err := c.Config.Substitute(); err != nil {
		return err
	}
	return nil
}

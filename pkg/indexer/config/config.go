package config

import (
	"github.com/baking-bad/noble-indexer/internal/profiler"
	"github.com/dipdup-net/go-lib/config"
)

type Config struct {
	config.Config    `yaml:",inline"`
	LogLevel         string           `validate:"omitempty,oneof=debug trace info warn error fatal panic" yaml:"log_level"`
	Indexer          Indexer          `yaml:"indexer"`
	API              API              `yaml:"api"`
	Profiler         *profiler.Config `validate:"omitempty"                                               yaml:"profiler"`
	MetadataResolver MetadataResolver `yaml:"metadata_resolver"`
}

type Indexer struct {
	Name            string `validate:"omitempty"       yaml:"name"`
	StartLevel      int64  `validate:"omitempty"       yaml:"start_level"`
	BlockPeriod     int64  `validate:"omitempty"       yaml:"block_period"`
	ScriptsDir      string `validate:"omitempty,dir"   yaml:"scripts_dir"`
	AssetsDir       string `validate:"omitempty,dir"   yaml:"assets_dir"`
	RequestBulkSize int    `validate:"omitempty,min=1" yaml:"request_bulk_size"`
}

type API struct {
	Bind           string `validate:"required" yaml:"bind"`
	RateLimit      int    `validate:"min=0"    yaml:"rate_limit"`
	RequestTimeout int    `validate:"min=1"    yaml:"request_timeout"`
}

type MetadataResolver struct {
	Name             string `validate:"omitempty" yaml:"name"`
	SyncPeriod       int64  `validate:"omitempty" yaml:"sync_period"`
	MetadataGateways string `validate:"omitempty" yaml:"metadata_gateways"`
}

// Substitute -
func (c *Config) Substitute() error {
	if err := c.Config.Substitute(); err != nil {
		return err
	}
	return nil
}

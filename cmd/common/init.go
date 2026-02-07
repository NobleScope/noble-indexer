package common

import (
	"os"
	"strconv"
	"time"

	"github.com/baking-bad/noble-indexer/internal/cache"
	"github.com/baking-bad/noble-indexer/internal/profiler"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	goLibConfig "github.com/dipdup-net/go-lib/config"
	"github.com/grafana/pyroscope-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	})
}

func InitConfig(rootCmd *cobra.Command) (*config.Config, error) {
	configPath := rootCmd.PersistentFlags().StringP("config", "c", "dipdup.yml", "path to YAML config file")
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("command line execute")
		return nil, err
	}

	if err := rootCmd.MarkFlagRequired("config"); err != nil {
		log.Panic().Err(err).Msg("config command line arg is required")
		return nil, err
	}

	var cfg config.Config
	if err := goLibConfig.Parse(*configPath, &cfg); err != nil {
		log.Panic().Err(err).Msg("parsing config file")
		return nil, err
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = zerolog.LevelInfoValue
	}

	return &cfg, nil
}

func InitLogger(level string) error {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Panic().Err(err).Msg("parsing log level")
		return err
	}
	zerolog.SetGlobalLevel(logLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}
	log.Logger = log.Logger.With().Caller().Logger()

	return nil
}

func InitProfiler(cfg *profiler.Config, serviceName string) (*pyroscope.Profiler, error) {
	return profiler.New(cfg, serviceName)
}

func InitCache(url string, ttl time.Duration) (cache.ICache, error) {
	if url != "" {
		return cache.NewValKey(url, ttl)
	}
	return nil, nil
}

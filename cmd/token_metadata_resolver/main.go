package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/baking-bad/noble-indexer/cmd/common"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/pkg/token_metadata"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "token_metadata_resolver",
	Short: "Noble | Token metadata resolver",
}

func main() {
	cfg, err := common.InitConfig(rootCmd)
	if err != nil {
		return
	}

	if err = common.InitLogger(cfg.LogLevel); err != nil {
		return
	}
	prscp, err := common.InitProfiler(cfg.Profiler, "token metadata resolver")
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	notifyCtx, notifyCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCancel()

	pg, err := postgres.Create(ctx, cfg.Database, cfg.Indexer.ScriptsDir)
	if err != nil {
		log.Panic().Err(err).Msg("can't create database connection")
		return
	}

	metadataResolver := token_metadata.NewModule(pg, *cfg)
	metadataResolver.Start(ctx)

	<-notifyCtx.Done()
	cancel()

	if err := metadataResolver.Close(); err != nil {
		log.Panic().Err(err).Msg("stopping metadata resolver")
	}

	if prscp != nil {
		if err := prscp.Stop(); err != nil {
			log.Panic().Err(err).Msg("stopping pyroscope")
		}
	}

	log.Info().Msg("stopped")
}

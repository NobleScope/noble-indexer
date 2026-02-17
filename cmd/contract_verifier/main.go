package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/NobleScope/noble-indexer/cmd/common"
	"github.com/NobleScope/noble-indexer/internal/storage/postgres"
	"github.com/NobleScope/noble-indexer/pkg/contract_verifier"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "contract_verifier",
	Short: "Noble | Contract verifier",
}

func main() {
	cfg, err := common.InitConfig(rootCmd)
	if err != nil {
		return
	}

	if err = common.InitLogger(cfg.LogLevel); err != nil {
		return
	}
	prscp, err := common.InitProfiler(cfg.Profiler, "contract verifier")
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	notifyCtx, notifyCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCancel()

	pg, err := postgres.Create(ctx, cfg.Database, cfg.Indexer.ScriptsDir, false)
	if err != nil {
		log.Panic().Err(err).Msg("can't create database connection")
		return
	}

	verifier := contract_verifier.NewModule(pg, *cfg)
	verifier.Start(ctx)

	<-notifyCtx.Done()
	cancel()

	if err := verifier.Close(); err != nil {
		log.Panic().Err(err).Msg("stopping metadata resolver")
	}

	if prscp != nil {
		if err := prscp.Stop(); err != nil {
			log.Panic().Err(err).Msg("stopping pyroscope")
		}
	}

	log.Info().Msg("stopped")
}

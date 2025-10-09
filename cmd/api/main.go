package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baking-bad/noble-indexer/cmd/common"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "Noble | API",
}

func main() {
	cfg, err := common.InitConfig(rootCmd)
	if err != nil {
		return
	}
	if err = common.InitLogger(cfg.LogLevel); err != nil {
		return
	}
	if err = run(cfg); err != nil {
		log.Fatal().Err(err).Msg("running API server")
	}
}

func run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifyCtx, notifyCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCancel()

	e := echo.New()
	e.HideBanner = true
	e.Use(zerologMiddleware())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	go func() {
		log.Info().Str("bind", cfg.API.Bind).Msg("Starting API server")
		if err := e.Start(cfg.API.Bind); err != nil {
			log.Err(err).Msg("API server error")
			cancel()
		}
	}()

	<-notifyCtx.Done()
	cancel()

	if err := e.Shutdown(context.Background()); err != nil {
		log.Panic().Err(err).Msg("stopping API server")
	}

	log.Info().Msg("stopped")
	return nil
}

func zerologMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			log.Info().
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Int("status", res.Status).
				Dur("latency", time.Since(start)).
				Str("remote_ip", c.RealIP()).
				Msg("request")

			return err
		}
	}
}

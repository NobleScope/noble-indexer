package main

import (
	"context"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/baking-bad/noble-indexer/cmd/api/docs"
	"github.com/baking-bad/noble-indexer/cmd/api/handler"
	"github.com/baking-bad/noble-indexer/cmd/common"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//	@title			Noble Indexer API
//	@version		1.0
//	@description	API for Noble blockchain indexer
//	@termsOfService	https://bakingbad.dev/terms

//	@contact.name	API Support
//	@contact.url	https://github.com/baking-bad/noble-indexer
//	@contact.email	hello@bakingbad.dev

//	@license.name	MIT
//	@license.url	https://github.com/baking-bad/noble-indexer/blob/master/LICENSE

//	@host		noble.dipdup.net
//	@BasePath	/v1

//	@schemes	https http

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
	e.Validator = handler.NewApiValidator()
	e.HideBanner = true
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogMethod:    true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			switch {
			case v.Status == http.StatusOK || v.Status == http.StatusNoContent:
				log.Info().
					Str("uri", v.URI).
					Int("status", v.Status).
					Dur("latency", v.Latency).
					Str("method", v.Method).
					Str("user_agent", v.UserAgent).
					Str("ip", c.RealIP()).
					Msg("request")
			case v.Status >= 500:
				log.Error().
					Str("uri", v.URI).
					Int("status", v.Status).
					Dur("latency", v.Latency).
					Str("method", v.Method).
					Str("user_agent", v.UserAgent).
					Str("ip", c.RealIP()).
					Msg("request")
			default:
				log.Warn().
					Str("uri", v.URI).
					Int("status", v.Status).
					Dur("latency", v.Latency).
					Str("method", v.Method).
					Str("user_agent", v.UserAgent).
					Str("ip", c.RealIP()).
					Msg("request")
			}
			return nil
		},
	}))
	e.Use(middleware.BodyLimit("9M"))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CORS())
	e.Pre(middleware.RemoveTrailingSlash())
	if cfg.API.RequestTimeout > 0 {
		e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: time.Duration(cfg.API.RequestTimeout) * time.Second,
		}))
	}
	if cfg.API.RateLimit > 0 {
		e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(cfg.API.RateLimit))))
	}

	db := initDatabase(cfg.Database, cfg.Indexer.ScriptsDir)
	initDispatcher(ctx, db)
	initHandlers(ctx, e, *cfg, db)

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
	if err := dispatcher.Close(); err != nil {
		log.Panic().Err(err).Msg("stopping dispatcher")
	}
	if err := db.Close(); err != nil {
		log.Panic().Err(err).Msg("closing database")
	}

	log.Info().Msg("stopped")
	return nil
}

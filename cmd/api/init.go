package main

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/cmd/api/handler"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	golibCfg "github.com/dipdup-net/go-lib/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func initDatabase(cfg golibCfg.Database, viewsDir string) postgres.Storage {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := postgres.Create(ctx, cfg, viewsDir)
	if err != nil {
		panic(err)
	}
	return db
}

func initHandlers(ctx context.Context, e *echo.Echo, cfg config.Config, db postgres.Storage) {
	v1 := e.Group("v1")

	stateHandlers := handler.NewStateHandler(db.State, cfg.Indexer.Name)
	v1.GET("/head", stateHandlers.Head)

	blockHandlers := handler.NewBlockHandler(db.Blocks, db.BlockStats, db.State, cfg.Indexer.Name)
	blockGroup := v1.Group("/block")
	{
		blockGroup.GET("", blockHandlers.List)
		blockGroup.GET("/count", blockHandlers.Count)
		heightGroup := blockGroup.Group("/:height")
		{
			heightGroup.GET("", blockHandlers.Get)
			heightGroup.GET("/stats", blockHandlers.GetStats)
		}
	}

	log.Info().Msg("API routes:")
	for _, route := range e.Routes() {
		log.Info().Msgf("[%s] %s -> %s", route.Method, route.Path, route.Name)
	}
}

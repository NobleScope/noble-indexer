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
	echoSwagger "github.com/swaggo/echo-swagger"
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
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := e.Group("v1")

	stateHandlers := handler.NewStateHandler(db.State, cfg.Indexer.Name)
	v1.GET("/head", stateHandlers.Head)

	constantsHandler := handler.NewConstantHandler()
	v1.GET("/enums", constantsHandler.Enums)

	blockHandlers := handler.NewBlockHandler(db.Blocks, db.BlockStats, db.Tx, db.State, cfg.Indexer.Name)
	blockGroup := v1.Group("/block")
	{
		blockGroup.GET("", blockHandlers.List)
		blockGroup.GET("/count", blockHandlers.Count)
		heightGroup := blockGroup.Group("/:height")
		{
			heightGroup.GET("", blockHandlers.Get)
			heightGroup.GET("/stats", blockHandlers.GetStats)
			heightGroup.GET("/transactions", blockHandlers.TransactionsList)
		}
	}
	txHandlers := handler.NewTxHandler(db.Tx, db.Trace, db.Addresses, cfg.Indexer.Name)
	txGroup := v1.Group("/tx")
	{
		txGroup.GET("", txHandlers.List)
		txGroup.GET("/:hash", txHandlers.Get)
	}
	v1.GET("/traces", txHandlers.Traces)

	logHandlers := handler.NewLogHandler(db.Logs, db.Tx, db.Addresses)
	v1.GET("/log", logHandlers.List)

	addressHandlers := handler.NewAddressHandler(db.Addresses)
	addressesGroup := v1.Group("/address")
	{
		addressesGroup.GET("", addressHandlers.List)
		addressGroup := addressesGroup.Group("/:hash")
		{
			addressGroup.GET("", addressHandlers.Get)
		}
	}

	contractHandlers := handler.NewContractHandler(db.Contracts, db.Tx, db.Sources)
	contractsGroup := v1.Group("/contract")
	{
		contractsGroup.GET("", contractHandlers.List)
		hashGroup := contractsGroup.Group("/:hash")
		{
			hashGroup.GET("", contractHandlers.Get)
			hashGroup.GET("/sources", contractHandlers.ContractSources)
		}
	}

	tokenHandlers := handler.NewTokenHandler(db.Token, db.Transfer, db.TokenBalance, db.Addresses)
	tokensGroup := v1.Group("/token")
	{
		tokensGroup.GET("", tokenHandlers.List)
		tokensGroup.GET("/:contract/:token_id", tokenHandlers.Get)
	}
	tokenTransfersGroup := v1.Group("/transfers")
	{
		tokenTransfersGroup.GET("", tokenHandlers.TransferList)
		tokenTransfersGroup.GET("/:id", tokenHandlers.GetTransfer)
	}
	v1.GET("/token_balance", tokenHandlers.TokenBalanceList)

	searchHandler := handler.NewSearchHandler(db.Search, db.Addresses, db.Blocks, db.Tx, db.Token)
	v1.GET("/search", searchHandler.Search)

	log.Info().Msg("API routes:")
	for _, route := range e.Routes() {
		log.Info().Msgf("[%s] %s -> %s", route.Method, route.Path, route.Name)
	}
}

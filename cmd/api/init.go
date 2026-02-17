package main

import (
	"context"
	"time"

	"github.com/NobleScope/noble-indexer/cmd/api/bus"
	apiCache "github.com/NobleScope/noble-indexer/cmd/api/cache"
	"github.com/NobleScope/noble-indexer/cmd/api/handler"
	"github.com/NobleScope/noble-indexer/cmd/api/handler/websocket"
	"github.com/NobleScope/noble-indexer/internal/cache"
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/postgres"
	"github.com/NobleScope/noble-indexer/pkg/indexer/config"
	golibCfg "github.com/dipdup-net/go-lib/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var dispatcher *bus.Dispatcher

func initDispatcher(ctx context.Context, db postgres.Storage) {
	d, err := bus.NewDispatcher(db)
	if err != nil {
		panic(err)
	}
	dispatcher = d
	dispatcher.Start(ctx)
}

func initDatabase(cfg golibCfg.Database, viewsDir string) postgres.Storage {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := postgres.Create(ctx, cfg, viewsDir, false)
	if err != nil {
		panic(err)
	}
	return db
}

func initHandlers(ctx context.Context, e *echo.Echo, cfg config.Config, db postgres.Storage, ttlCache cache.ICache) {
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	defaultMiddlewareCache := apiCache.Middleware(ttlCache, nil, nil)

	v1 := e.Group("v1")

	stateHandlers := handler.NewStateHandler(db.State, cfg.Indexer.Name)
	v1.GET("/head", stateHandlers.Head)

	constantsHandler := handler.NewConstantHandler()
	v1.GET("/enums", constantsHandler.Enums, defaultMiddlewareCache)

	blockHandlers := handler.NewBlockHandler(db.Blocks, db.BlockStats, db.Tx, db.State, cfg.Indexer.Name)
	blockGroup := v1.Group("/blocks")
	{
		blockGroup.GET("", blockHandlers.List)
		blockGroup.GET("/count", blockHandlers.Count)
		heightGroup := blockGroup.Group("/:height")
		{
			heightGroup.GET("", blockHandlers.Get, defaultMiddlewareCache)
			heightGroup.GET("/stats", blockHandlers.GetStats, defaultMiddlewareCache)
			heightGroup.GET("/transactions", blockHandlers.TransactionsList, defaultMiddlewareCache)
		}
	}
	txHandlers := handler.NewTxHandler(db.Tx, db.Trace, db.Addresses, cfg.Indexer.Name)
	txGroup := v1.Group("/txs")
	{
		txGroup.GET("", txHandlers.List)
		txGroup.GET("/:hash", txHandlers.Get, defaultMiddlewareCache)
		txGroup.GET("/:hash/traces_tree", txHandlers.TxTracesTree, defaultMiddlewareCache)
	}
	v1.GET("/traces", txHandlers.Traces)

	logHandlers := handler.NewLogHandler(db.Logs, db.Tx, db.Addresses)
	v1.GET("/logs", logHandlers.List)

	addressHandlers := handler.NewAddressHandler(db.Addresses)
	addressesGroup := v1.Group("/addresses")
	{
		addressesGroup.GET("", addressHandlers.List)
		addressGroup := addressesGroup.Group("/:hash")
		{
			addressGroup.GET("", addressHandlers.Get)
		}
	}

	contractHandlers := handler.NewContractHandler(db.Contracts, db.Addresses, db.Tx, db.Sources)
	contractsGroup := v1.Group("/contracts")
	{
		contractsGroup.GET("", contractHandlers.List)
		hashGroup := contractsGroup.Group("/:hash")
		{
			hashGroup.GET("", contractHandlers.Get)
			hashGroup.GET("/sources", contractHandlers.ContractSources, defaultMiddlewareCache)
			hashGroup.GET("/code", contractHandlers.GetCode, defaultMiddlewareCache)
		}
	}

	tokenHandlers := handler.NewTokenHandler(db.Token, db.Transfer, db.TokenBalance, db.Addresses, db.Tx)
	tokensGroup := v1.Group("/tokens")
	{
		tokensGroup.GET("", tokenHandlers.List)
		tokensGroup.GET("/:contract/:token_id", tokenHandlers.Get)
	}
	tokenTransfersGroup := v1.Group("/transfers")
	{
		tokenTransfersGroup.GET("", tokenHandlers.TransferList)
		tokenTransfersGroup.GET("/:id", tokenHandlers.GetTransfer)
	}
	v1.GET("/token_balances", tokenHandlers.TokenBalanceList)

	userOpHandlers := handler.NewUserOpHandler(db.ERC4337UserOps, db.Tx, db.Addresses)
	userOpsGroup := v1.Group("/user_ops")
	{
		userOpsGroup.GET("", userOpHandlers.List)
		userOpsGroup.GET("/:hash", userOpHandlers.Get, defaultMiddlewareCache)
	}

	searchHandler := handler.NewSearchHandler(db.Search, db.Addresses, db.Blocks, db.Tx, db.Token)
	v1.GET("/search", searchHandler.Search)

	proxyHandlers := handler.NewProxyContractHandler(db.ProxyContracts, db.Addresses, cfg.Indexer.Name)
	proxyGroup := v1.Group("/proxy")
	{
		proxyGroup.GET("", proxyHandlers.List)
	}

	statsHandler := handler.NewStatsHandler(db.State, db.BlockStats, cfg.Indexer.Name)
	statsGroup := v1.Group("/stats")
	{
		statsGroup.GET("/block_time", statsHandler.AvgBlockTime, defaultMiddlewareCache)
	}

	beaconWithdrawalHandler := handler.NewBeaconWithdrawalHandler(db.BeaconWithdrawal, db.Addresses)
	beaconWithdrawalsGroup := v1.Group("/beacon_withdrawals")
	{
		beaconWithdrawalsGroup.GET("", beaconWithdrawalHandler.List)
	}

	contractVerificationHandler := handler.NewContractVerificationHandler(db.Contracts, db.VerificationTasks, db.VerificationFiles, db.Transactable)
	verificationGroup := v1.Group("/verification/code")
	{
		verificationGroup.POST("", contractVerificationHandler.ContractVerify)
	}

	if cfg.API.Websocket {
		initWebsocket(ctx, v1)
	}

	log.Info().Msg("API routes:")
	for _, route := range e.Routes() {
		log.Info().Msgf("[%s] %s -> %s", route.Method, route.Path, route.Name)
	}
}

var (
	wsManager *websocket.Manager
)

func initWebsocket(ctx context.Context, group *echo.Group) {
	observer := dispatcher.Observe(storage.ChannelHead, storage.ChannelBlock)
	wsManager = websocket.NewManager(observer)
	wsManager.Start(ctx)
	group.GET("/ws", wsManager.Handle)
}

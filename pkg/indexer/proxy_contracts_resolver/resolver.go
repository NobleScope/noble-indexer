package proxy_contracts_resolver

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/postgres"
	"github.com/NobleScope/noble-indexer/pkg/indexer/config"
	"github.com/NobleScope/noble-indexer/pkg/node/rpc"
	"github.com/dipdup-io/workerpool"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	sdkSync "github.com/dipdup-net/indexer-sdk/pkg/sync"
)

type Module struct {
	modules.BaseModule

	cfg       config.Indexer
	api       *rpc.API
	pg        *postgres.Storage
	pool      *workerpool.TimedPool[[]*storage.ProxyContract]
	taskQueue *sdkSync.Map[string, struct{}]
}

var _ modules.Module = (*Module)(nil)

const (
	OutputName = "data"
	StopOutput = "stop"
)

func NewModule(cfg config.Indexer, api *rpc.API, pg *postgres.Storage) Module {
	m := Module{
		BaseModule: modules.New("proxy_contracts_resolver"),
		cfg:        cfg,
		api:        api,
		pg:         pg,
		taskQueue:  sdkSync.NewMap[string, struct{}](),
	}

	m.pool = workerpool.NewTimedPool(
		m.dispatcher,
		m.worker,
		m.dispatcherErrorHandler,
		cfg.Proxy.Threads,
		cfg.Proxy.SyncPeriodSeconds*1000,
	)

	m.CreateOutput(OutputName)
	m.CreateOutput(StopOutput)

	return m
}

func (p *Module) Start(ctx context.Context) {
	p.Log.Info().Msg("starting...")
	p.G.GoCtx(ctx, p.pool.Start)
}

func (p *Module) Close() error {
	p.Log.Info().Msg("closing...")
	p.G.Wait()
	return nil
}

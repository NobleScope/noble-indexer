package receiver

import (
	"context"
	"sync"

	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/baking-bad/noble-indexer/pkg/node"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	"github.com/gorilla/websocket"
)

const (
	BlocksOutput = "blocks"
	StopOutput   = "stop"
)

type Module struct {
	modules.BaseModule
	api              node.Api
	ws               *websocket.Conn
	level            types.Level
	blocks           chan types.BlockData
	hash             string
	mx               *sync.RWMutex
	cfg              config.Indexer
	cancelReadBlocks context.CancelFunc
	w                *Worker
}

var _ modules.Module = (*Module)(nil)

func NewModule(cfg config.Indexer, api node.Api, ws *websocket.Conn) Module {
	level := types.Level(cfg.StartLevel)
	receiver := Module{
		BaseModule: modules.New("receiver"),
		api:        api,
		cfg:        cfg,
		ws:         ws,
		level:      level,
		blocks:     make(chan types.BlockData, 128),
		mx:         new(sync.RWMutex),
	}

	receiver.w = NewWorker(api, receiver.Log, receiver.blocks, cfg.RequestBulkSize)

	receiver.CreateOutput(BlocksOutput)
	receiver.CreateOutput(StopOutput)

	return receiver
}

func (r *Module) Start(ctx context.Context) {
	r.Log.Info().Msg("starting receiver...")

	r.G.GoCtx(ctx, r.sequencer)
	r.G.GoCtx(ctx, r.sync)
}

func (r *Module) Level() (types.Level, string) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.level, r.hash
}

func (r *Module) setLevel(level types.Level, hash string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.level = level
	r.hash = hash
}

func (r *Module) Close() error {
	r.Log.Info().Msg("closing...")
	r.G.Wait()

	close(r.blocks)

	return nil
}

func (r *Module) stopAll() {
	r.MustOutput(StopOutput).Push(struct{}{})
}

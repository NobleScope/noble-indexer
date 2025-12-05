package receiver

import (
	"context"
	"sync"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/baking-bad/noble-indexer/pkg/node"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	BlocksOutput     = "blocks"
	RollbackOutput   = "signal"
	RollbackInput    = "state"
	GenesisOutput    = "genesis"
	GenesisDoneInput = "genesis_done"
	StopOutput       = "stop"
)

type Module struct {
	modules.BaseModule
	api              node.Api
	ws               *websocket.Conn
	level            types.Level
	blocks           chan types.BlockData
	hash             []byte
	needGenesis      bool
	mx               *sync.RWMutex
	cfg              config.Indexer
	cancelReadBlocks context.CancelFunc
	w                *Worker
	rollbackSync     *sync.WaitGroup
}

var _ modules.Module = (*Module)(nil)

func NewModule(cfg config.Indexer, api node.Api, ws *websocket.Conn, state *storage.State) Module {
	level := types.Level(cfg.StartLevel)
	var lastHash []byte
	if state != nil {
		level = state.LastHeight
		lastHash = state.LastHash
	}

	receiver := Module{
		BaseModule:   modules.New("receiver"),
		api:          api,
		cfg:          cfg,
		ws:           ws,
		level:        level,
		hash:         lastHash,
		needGenesis:  state == nil,
		blocks:       make(chan types.BlockData, 128),
		mx:           new(sync.RWMutex),
		rollbackSync: new(sync.WaitGroup),
	}

	receiver.w = NewWorker(api, receiver.Log, receiver.blocks, cfg.RequestBulkSize)

	receiver.CreateInput(RollbackInput)
	receiver.CreateInput(GenesisDoneInput)

	receiver.CreateOutput(BlocksOutput)
	receiver.CreateOutput(RollbackOutput)
	receiver.CreateOutput(GenesisOutput)
	receiver.CreateOutput(StopOutput)

	return receiver
}

func (r *Module) Start(ctx context.Context) {
	r.Log.Info().Msg("starting receiver...")

	if r.needGenesis {
		if err := r.receiveGenesis(ctx); err != nil {
			log.Err(err).Msg("receiving genesis error")
			return
		}
	}

	r.G.GoCtx(ctx, r.sequencer)
	r.G.GoCtx(ctx, r.sync)
	r.G.GoCtx(ctx, r.rollback)
}

func (r *Module) Level() (types.Level, []byte) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.level, r.hash
}

func (r *Module) setLevel(level types.Level, hash []byte) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.level = level
	r.hash = hash
}

func (r *Module) rollback(ctx context.Context) {
	rollbackInput := r.MustInput(RollbackInput)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-rollbackInput.Listen():
			if !ok {
				r.Log.Warn().Msg("can't read message from rollback input, channel is closed and drained")
				continue
			}

			state, ok := msg.(storage.State)
			if !ok {
				r.Log.Warn().Msgf("invalid message type: %T", msg)
				continue
			}

			r.w.queue = r.w.queue[:0]
			r.setLevel(state.LastHeight, state.LastHash)
			r.Log.Info().Msgf("caught return from rollback to level=%d", state.LastHeight)
			r.rollbackSync.Done()
		}
	}
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

package storage

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	decodeContext "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

const (
	InputName  = "data"
	StopOutput = "stop"
)

type Module struct {
	modules.BaseModule
	pg          postgres.Storage
	storage     sdk.Transactable
	indexerName string
}

var _ modules.Module = (*Module)(nil)

func NewModule(
	pg postgres.Storage,
	cfg config.Indexer,
) Module {
	m := Module{
		BaseModule:  modules.New("storage"),
		pg:          pg,
		storage:     pg.Transactable,
		indexerName: cfg.Name,
	}

	m.CreateInputWithCapacity(InputName, 128)
	m.CreateOutput(StopOutput)

	return m
}

func (module *Module) Start(ctx context.Context) {
	module.G.GoCtx(ctx, module.listen)
}

func (module *Module) listen(ctx context.Context) {
	module.Log.Info().Msg("module started")
	input := module.MustInput(InputName)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-input.Listen():
			if !ok {
				module.Log.Warn().Msg("can't read message from input")
				module.MustOutput(StopOutput).Push(struct{}{})
				continue
			}

			decodedContext, ok := msg.(*decodeContext.Context)
			if !ok {
				module.Log.Warn().Msgf("invalid message type: %T", msg)
				continue
			}

			state, err := module.saveBlock(ctx, decodedContext)
			if err != nil {
				module.Log.Err(err).
					Uint64("height", uint64(decodedContext.Block.Height)).
					Any("state", state).
					Msg("block saving error")
				module.MustOutput(StopOutput).Push(struct{}{})
				continue
			}
		}
	}
}

// Close -
func (module *Module) Close() error {
	module.Log.Info().Msg("closing module...")
	module.G.Wait()
	return nil
}

func (module *Module) saveBlock(ctx context.Context, dCtx *decodeContext.Context) (storage.State, error) {
	start := time.Now()
	module.Log.Info().Uint64("height", uint64(dCtx.Block.Height)).Msg("saving block...")
	tx, err := postgres.BeginTransaction(ctx, module.storage)
	if err != nil {
		return storage.State{}, err
	}
	defer tx.Close(ctx)

	state, err := module.processBlockInTransaction(ctx, tx, dCtx)
	if err != nil {
		return state, tx.HandleError(ctx, err)
	}

	if err := tx.Flush(ctx); err != nil {
		return state, tx.HandleError(ctx, err)
	}
	module.Log.Info().
		Uint64("height", uint64(dCtx.Block.Height)).
		Time("block_time", dCtx.Block.Time).
		Int64("ms", time.Since(start).Milliseconds()).
		Int("tx_count", len(dCtx.Block.Txs)).
		Msg("block saved")
	return state, nil
}

func (module *Module) processBlockInTransaction(ctx context.Context, tx storage.Transaction, dCtx *decodeContext.Context) (storage.State, error) {
	block := dCtx.Block
	state, err := module.pg.State.ByName(ctx, module.indexerName)
	if err != nil {
		if module.pg.State.IsNoRows(err) {
			if err = tx.Add(ctx, &storage.State{
				Name:          module.indexerName,
				LastHeight:    block.Height,
				LastTime:      block.Time,
				LastHash:      block.Hash,
				TotalTx:       int64(len(block.Txs)),
				TotalAccounts: int64(len(dCtx.GetAddresses())),
			}); err != nil {
				return state, err
			}
		} else {
			return state, err
		}
	}

	// todo: handle genesis block

	if err := tx.Add(ctx, block); err != nil {
		return state, err
	}

	if err := tx.Add(ctx, block.Stats); err != nil {
		return state, err
	}

	addrToId, totalAccounts, err := saveAddresses(ctx, tx, dCtx.GetAddresses())
	if err != nil {
		return state, err
	}

	err = saveTransactions(ctx, tx, block.Txs, addrToId)
	if err != nil {
		return state, err
	}

	txHashToId := make(map[string]uint64)
	for i := range block.Txs {
		txHashToId[block.Txs[i].Hash.String()] = block.Txs[i].Id
	}

	contractToId, err := saveContracts(ctx, tx, dCtx.GetContracts(), txHashToId, addrToId)
	if err != nil {
		return state, err
	}

	err = saveTraces(ctx, tx, dCtx.GetTraces(), txHashToId, addrToId, contractToId)
	if err != nil {
		return state, err
	}

	logs := make([]storage.Log, 0, 10000)
	for i := range block.Txs {
		for j := range block.Txs[i].Logs {
			block.Txs[i].Logs[j].TxId = block.Txs[i].Id
		}

		logs = append(logs, block.Txs[i].Logs...)
		if len(logs) >= 10000 {
			if err := tx.SaveLogs(ctx, logs...); err != nil {
				return state, err
			}
			logs = make([]storage.Log, 0, 10000)
		}
	}
	if len(logs) > 0 {
		if err := tx.SaveLogs(ctx, logs...); err != nil {
			return state, err
		}
	}

	if err := updateState(block, totalAccounts, int64(len(block.Txs)), &state); err != nil {
		return state, err
	}

	err = tx.Update(ctx, &state)

	return state, err
}

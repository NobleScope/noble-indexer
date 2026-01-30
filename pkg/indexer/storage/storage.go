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
	"github.com/goccy/go-json"
)

const (
	InputName           = "data"
	ProxyContractsInput = "proxy_contracts"
	StopOutput          = "stop"
)

type Module struct {
	modules.BaseModule
	pg          postgres.Storage
	storage     sdk.Transactable
	notificator storage.Notificator
	indexerName string
}

var _ modules.Module = (*Module)(nil)

func NewModule(
	pg postgres.Storage,
	notificator storage.Notificator,
	cfg config.Indexer,
) Module {
	m := Module{
		BaseModule:  modules.New("storage"),
		pg:          pg,
		storage:     pg.Transactable,
		notificator: notificator,
		indexerName: cfg.Name,
	}

	m.CreateInputWithCapacity(InputName, 128)
	m.CreateInputWithCapacity(ProxyContractsInput, 128)
	m.CreateOutput(StopOutput)

	return m
}

func (module *Module) Start(ctx context.Context) {
	module.G.GoCtx(ctx, module.listen)
	module.G.GoCtx(ctx, module.listenProxyImplementations)
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

			if err := module.notify(ctx, state, *decodedContext.Block); err != nil {
				module.Log.Err(err).Msg("block notification error")
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

func (module *Module) processBlockInTransaction(
	ctx context.Context,
	tx storage.Transaction,
	dCtx *decodeContext.Context,
) (storage.State, error) {
	block := dCtx.Block
	state, err := module.pg.State.ByName(ctx, module.indexerName)
	if err != nil {
		return state, err
	}
	if state.LastHeight > 0 {
		block.Stats.BlockTime = uint64(block.Time.Sub(state.LastTime).Milliseconds())
	}

	addrToId, totalAccounts, err := saveAddresses(ctx, tx, dCtx.GetAddresses())
	if err != nil {
		return state, err
	}

	err = saveBlock(ctx, tx, block, addrToId)
	if err != nil {
		return state, err
	}

	if err := tx.Add(ctx, block.Stats); err != nil {
		return state, err
	}

	err = saveTransactions(ctx, tx, block.Txs, addrToId)
	if err != nil {
		return state, err
	}

	txHashToId := make(map[string]uint64)
	transfers := make([]*storage.Transfer, 0)
	for i := range block.Txs {
		txHashToId[block.Txs[i].Hash.String()] = block.Txs[i].Id

		for j := range block.Txs[i].Transfers {
			block.Txs[i].Transfers[j].TxID = block.Txs[i].Id
		}
		transfers = append(transfers, block.Txs[i].Transfers...)
	}

	totalContracts, err := saveContracts(ctx, tx, dCtx.GetContracts(), txHashToId, addrToId)
	if err != nil {
		return state, err
	}

	err = saveTraces(ctx, tx, dCtx.GetTraces(), txHashToId, addrToId)
	if err != nil {
		return state, err
	}

	err = saveLogs(ctx, tx, block.Txs, addrToId)
	if err != nil {
		return state, err
	}

	err = saveTransfers(ctx, tx, transfers, addrToId)
	if err != nil {
		return state, err
	}

	totalTokens, err := saveTokens(ctx, tx, dCtx.GetTokens(), addrToId)
	if err != nil {
		return state, err
	}

	err = saveTokenBalances(ctx, tx, dCtx.GetTokenBalances(), addrToId)
	if err != nil {
		return state, err
	}

	err = saveProxyContracts(ctx, tx, dCtx.GetProxyContracts(), addrToId)
	if err != nil {
		return state, err
	}

	err = saveERC4337UserOps(ctx, tx, dCtx.GetUserOps(), txHashToId, addrToId)
	if err != nil {
		return state, err
	}

	if err := updateState(block, totalAccounts, int64(len(block.Txs)), totalContracts, 0, totalTokens, &state); err != nil {
		return state, err
	}

	err = tx.Update(ctx, &state)

	return state, err
}

func (module *Module) notify(ctx context.Context, state storage.State, block storage.Block) error {
	if time.Since(block.Time) > time.Hour {
		// do not notify all about events if initial indexing is in progress
		return nil
	}

	rawState, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := module.notificator.Notify(ctx, storage.ChannelHead, string(rawState)); err != nil {
		return err
	}

	rawBlock, err := json.Marshal(block)
	if err != nil {
		return err
	}
	if err := module.notificator.Notify(ctx, storage.ChannelBlock, string(rawBlock)); err != nil {
		return err
	}

	return nil
}

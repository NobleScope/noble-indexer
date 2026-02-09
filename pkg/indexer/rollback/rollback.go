package rollback

import (
	"bytes"
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/baking-bad/noble-indexer/pkg/node"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	InputName  = "signal"
	OutputName = "state"
	StopOutput = "stop"
)

type Module struct {
	modules.BaseModule
	tx        sdk.Transactable
	state     storage.IState
	blocks    storage.IBlock
	node      node.Api
	indexName string
}

var _ modules.Module = (*Module)(nil)

func NewModule(
	tx sdk.Transactable,
	state storage.IState,
	blocks storage.IBlock,
	node node.Api,
	cfg config.Indexer,
) Module {
	module := Module{
		BaseModule: modules.New("rollback"),
		tx:         tx,
		state:      state,
		blocks:     blocks,
		node:       node,
		indexName:  cfg.Name,
	}

	module.CreateInput(InputName)
	module.CreateOutput(OutputName)
	module.CreateOutput(StopOutput)

	return module
}

// Start -
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
		case _, ok := <-input.Listen():
			if !ok {
				module.Log.Warn().Msg("can't read message from input, channel was dried and closed")
				module.MustOutput(StopOutput).Push(struct{}{})
				return
			}

			if err := module.rollback(ctx); err != nil {
				module.Log.Err(err).Msgf("error occurred")
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

func (module *Module) rollback(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			lastBlock, err := module.blocks.Last(ctx)
			if err != nil {
				return errors.Wrap(err, "receive last block from database")
			}

			nodeBlock, err := module.node.Block(ctx, lastBlock.Height)
			if err != nil {
				return errors.Wrapf(err, "receive block from node by height: %d", lastBlock.Height)
			}

			log.Debug().
				Uint64("height", uint64(lastBlock.Height)).
				Hex("db_block_hash", lastBlock.Hash).
				Hex("node_block_hash", nodeBlock.Hash).
				Msg("comparing hash...")

			if bytes.Equal(lastBlock.Hash, nodeBlock.Hash) {
				return module.finish(ctx)
			}

			log.Warn().
				Uint64("height", uint64(lastBlock.Height)).
				Hex("db_block_hash", lastBlock.Hash).
				Hex("node_block_hash", nodeBlock.Hash).
				Msg("need rollback")

			if err := module.rollbackBlock(ctx, lastBlock); err != nil {
				return errors.Wrapf(err, "rollback block: %d", lastBlock.Height)
			}
		}
	}
}

func (module *Module) finish(ctx context.Context) error {
	newState, err := module.state.ByName(ctx, module.indexName)
	if err != nil {
		return err
	}
	module.MustOutput(OutputName).Push(newState)

	log.Info().
		Uint64("new_height", uint64(newState.LastHeight)).
		Msg("roll backed to new height")

	return nil
}

func (module *Module) rollbackBlock(ctx context.Context, block storage.Block) error {
	tx, err := postgres.BeginTransaction(ctx, module.tx)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	height := block.Height
	if err := tx.RollbackBlock(ctx, height); err != nil {
		return tx.HandleError(ctx, err)
	}
	blockStats, err := tx.RollbackBlockStats(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	addresses, err := tx.RollbackAddresses(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	txs, err := tx.RollbackTxs(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	err = tx.RollbackContracts(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	err = tx.RollbackLogs(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	traces, err := tx.RollbackTraces(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	transfers, err := tx.RollbackTransfers(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	tokens, err := tx.RollbackTokens(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	err = tx.RollbackERC4337UserOps(ctx, height)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	if err := tx.RollbackBeaconWithdrawals(ctx, height); err != nil {
		return tx.HandleError(ctx, err)
	}

	if err := module.rollbackBalances(ctx, tx, block, txs, traces, transfers, tokens, addresses); err != nil {
		return tx.HandleError(ctx, err)
	}

	newBlock, err := tx.LastBlock(ctx)
	if err != nil {
		return tx.HandleError(ctx, err)
	}
	state, err := tx.State(ctx, module.indexName)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	state.LastHeight = newBlock.Height
	state.LastHash = newBlock.Hash
	state.LastTime = newBlock.Time
	state.TotalTx -= blockStats.TxCount
	state.TotalAccounts -= int64(len(addresses))

	if err := tx.Update(ctx, &state); err != nil {
		return tx.HandleError(ctx, err)
	}

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	log.Warn().
		Uint64("height", uint64(newBlock.Height)).
		Hex("block_hash", newBlock.Hash).
		Msg("rollback completed")

	return nil
}

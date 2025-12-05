package receiver

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/baking-bad/noble-indexer/pkg/types"
)

func (r *Module) sequencer(ctx context.Context) {
	orderedBlocks := map[uint64]types.BlockData{}
	l, prevBlockHash := r.Level()
	currentBlock := uint64(l + 1)

	for {
		select {
		case <-ctx.Done():
			return
		case block, ok := <-r.blocks:
			if !ok {
				r.Log.Warn().Msg("can't read message from blocks input, channel was dried and closed")
				r.stopAll()
				return
			}

			blockNumber, err := block.Number.Uint64()
			if err != nil {
				r.Log.Error().Err(err).Msg("can't read block number")
			}
			orderedBlocks[blockNumber] = block

			b, ok := orderedBlocks[currentBlock]
			for ok {
				if prevBlockHash != nil {
					if !bytes.Equal(b.ParentHash, prevBlockHash) {
						prevBlockHash, currentBlock, orderedBlocks = r.startRollback(b, prevBlockHash)
						break
					}
				}

				r.MustOutput(BlocksOutput).Push(b)
				r.setLevel(types.Level(currentBlock), b.Hash)
				r.Log.Debug().
					Uint64("height", currentBlock).
					Msg("put in order block")

				prevBlockHash = b.Hash
				delete(orderedBlocks, currentBlock)
				currentBlock += 1

				b, ok = orderedBlocks[currentBlock]
			}
		}
	}
}

func (r *Module) startRollback(
	b types.BlockData,
	prevBlockHash []byte,
) ([]byte, uint64, map[uint64]types.BlockData) {
	blockNumber, err := b.Number.Uint64()
	if err != nil {
		r.Log.Error().Err(err).Msg("can't read block number")
	}
	r.Log.Info().
		Str("current.lastBlockHash", hex.EncodeToString(b.Block.Hash)).
		Str("prevBlockHash", hex.EncodeToString(prevBlockHash)).
		Uint64("level", blockNumber).
		Msg("rollback detected")

	// Pause all receiver routines
	r.rollbackSync.Add(1)

	// Stop readBlocks
	if r.cancelReadBlocks != nil {
		r.cancelReadBlocks()
	}

	clearChannel(r.blocks)

	// Start rollback
	r.MustOutput(RollbackOutput).Push(struct{}{})

	// Wait until rollback will be finished
	r.rollbackSync.Wait()

	// Reset empty state
	level, hash := r.Level()
	currentBlock := uint64(level)
	prevBlockHash = hash
	orderedBlocks := map[uint64]types.BlockData{}

	return prevBlockHash, currentBlock, orderedBlocks
}

func clearChannel(blocks <-chan types.BlockData) {
	for len(blocks) > 0 {
		<-blocks
	}
}

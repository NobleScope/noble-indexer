package receiver

import (
	"context"

	"github.com/baking-bad/noble-indexer/pkg/types"
)

func (r *Module) sequencer(ctx context.Context) {
	orderedBlocks := map[uint64]types.BlockData{}
	l, _ := r.Level()
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
				r.MustOutput(BlocksOutput).Push(b)
				r.setLevel(types.Level(currentBlock), b.Hash.String())
				r.Log.Debug().
					Uint64("height", currentBlock).
					Msg("put in order block")

				delete(orderedBlocks, currentBlock)
				currentBlock += 1

				b, ok = orderedBlocks[currentBlock]
			}
		}
	}
}

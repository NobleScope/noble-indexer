package storage

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/pkg/errors"
)

func updateState(block *storage.Block, totalAccounts, totalTxs int64, state *storage.State) error {
	if block.Height <= state.LastHeight {
		return errors.Errorf("block has already indexed: height=%d  state=%d", block.Height, state.LastHeight)
	}

	state.LastHeight = block.Height
	state.LastHash = block.Hash
	state.LastTime = block.Time
	state.TotalTx += totalTxs
	state.TotalAccounts += totalAccounts
	return nil
}

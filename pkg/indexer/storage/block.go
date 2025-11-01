package storage

import (
	"context"
	"fmt"

	"github.com/baking-bad/noble-indexer/internal/storage"
)

func saveBlock(
	ctx context.Context,
	tx storage.Transaction,
	block *storage.Block,
	addresses map[string]uint64,
) error {
	if id, ok := addresses[block.Miner.Address]; ok {
		block.MinerId = id
		return tx.Add(ctx, block)
	}

	return fmt.Errorf("miner address %s not found in block", block.Miner.Address)
}

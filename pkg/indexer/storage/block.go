package storage

import (
	"context"
	"fmt"

	"github.com/NobleScope/noble-indexer/internal/storage"
)

func saveBlock(
	ctx context.Context,
	tx storage.Transaction,
	block *storage.Block,
	addresses map[string]uint64,
) error {
	if id, ok := addresses[block.Miner.Hash.String()]; ok {
		block.MinerId = id
		return tx.Add(ctx, block)
	}

	return fmt.Errorf("miner address %s not found in block", block.Miner.Hash.String())
}

package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
)

func saveAddresses(
	ctx context.Context,
	tx storage.Transaction,
	addresses []*storage.Address,
) (map[string]uint64, int64, error) {
	if len(addresses) == 0 {
		return nil, 0, nil
	}

	totalAccounts, err := tx.SaveAddresses(ctx, addresses...)
	if err != nil {
		return nil, 0, err
	}

	addToId := make(map[string]uint64)
	for i := range addresses {
		addToId[addresses[i].Address] = addresses[i].Id
	}
	return addToId, totalAccounts, err
}

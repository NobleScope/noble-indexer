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

	addrToId := make(map[string]uint64)
	balances := make([]*storage.Balance, 0)
	for i := range addresses {
		addrToId[addresses[i].Address] = addresses[i].Id
		addresses[i].Balance.Id = addresses[i].Id
		balances = append(balances, addresses[i].Balance)
	}
	if len(balances) > 0 {
		err = tx.SaveBalances(ctx, balances...)
	}

	return addrToId, totalAccounts, err
}

package storage

import (
	"context"
	"fmt"

	"github.com/NobleScope/noble-indexer/internal/storage"
)

func saveBeaconWithdrawals(
	ctx context.Context,
	tx storage.Transaction,
	withdrawals []*storage.BeaconWithdrawal,
	addresses map[string]uint64,
) error {
	if len(withdrawals) == 0 {
		return nil
	}

	for i := range withdrawals {
		id, ok := addresses[withdrawals[i].Address.Hash.String()]
		if !ok {
			return fmt.Errorf("withdrawal address %s not found", withdrawals[i].Address.Hash.String())
		}
		withdrawals[i].AddressId = id
	}

	return tx.SaveBeaconWithdrawals(ctx, withdrawals...)
}

package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/pkg/errors"
)

func saveTokenBalances(
	ctx context.Context,
	tx storage.Transaction,
	tokenBalances []*storage.TokenBalance,
	addresses map[string]uint64,
) error {
	if len(tokenBalances) == 0 {
		return nil
	}

	for i := range tokenBalances {
		contractID, ok := addresses[tokenBalances[i].Contract.Address]
		if !ok {
			return errors.Errorf("can't find contract key: %s", tokenBalances[i].Contract.Address)
		}
		tokenBalances[i].ContractID = contractID

		addressID, ok := addresses[tokenBalances[i].Address.Address]
		if !ok {
			return errors.Errorf("can't find address key: %s", tokenBalances[i].Address.Address)
		}
		tokenBalances[i].AddressID = addressID
	}

	err := tx.SaveTokenBalances(ctx, tokenBalances...)
	return err
}

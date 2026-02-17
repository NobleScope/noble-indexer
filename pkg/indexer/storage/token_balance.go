package storage

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
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
		contractID, ok := addresses[tokenBalances[i].Contract.Address.String()]
		if !ok {
			return errors.Errorf("can't find contract key: %s", tokenBalances[i].Contract.Address.String())
		}
		tokenBalances[i].ContractID = contractID

		addressID, ok := addresses[tokenBalances[i].Address.String()]
		if !ok {
			return errors.Errorf("can't find address key: %s", tokenBalances[i].Address.String())
		}
		tokenBalances[i].AddressID = addressID
	}

	_, err := tx.SaveTokenBalances(ctx, tokenBalances...)
	return err
}

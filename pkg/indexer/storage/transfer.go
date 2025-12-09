package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/pkg/errors"
)

func saveTransfers(
	ctx context.Context,
	tx storage.Transaction,
	transfers []*storage.Transfer,
	addresses map[string]uint64,
) error {
	if len(transfers) == 0 {
		return nil
	}

	for i := range transfers {
		contractID, ok := addresses[transfers[i].Contract.Address.Address]
		if !ok {
			return errors.Errorf("can't find contract key: %s", transfers[i].Contract.Address.Address)
		}
		transfers[i].ContractId = contractID

		if transfers[i].FromAddress != nil {
			id, ok := addresses[transfers[i].FromAddress.Address]
			if !ok {
				return errors.Errorf("can't find from address key: %s", transfers[i].FromAddress.Address)
			}
			transfers[i].FromAddressId = &id
		}

		if transfers[i].ToAddress != nil {
			id, ok := addresses[transfers[i].ToAddress.Address]
			if !ok {
				return errors.Errorf("can't find to address key: %s", transfers[i].ToAddress.Address)
			}
			transfers[i].ToAddressId = &id
		}
	}

	err := tx.SaveTransfers(ctx, transfers...)
	return err
}

package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/pkg/errors"
)

func saveContracts(
	ctx context.Context,
	tx storage.Transaction,
	contracts []*storage.Contract,
	txHashes map[string]uint64,
	addresses map[string]uint64,
) (int64, error) {
	if len(contracts) == 0 {
		return 0, nil
	}

	for _, contract := range contracts {
		id, ok := addresses[contract.Address.Hash.String()]
		if !ok {
			return 0, errors.Errorf("can't find contract key: %s", contract.Address.Hash.String())
		}
		contract.Id = id
		if contract.Tx == nil {
			continue
		}

		txId, ok := txHashes[contract.Tx.Hash.String()]
		if !ok {
			return 0, errors.Errorf("can't find tx hash: %s", contract.Tx.Hash.String())
		}
		contract.TxId = &txId
	}

	totalContracts, err := tx.SaveContracts(ctx, contracts...)
	if err != nil {
		return 0, err
	}

	return totalContracts, err
}

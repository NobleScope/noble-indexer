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
) (map[string]uint64, error) {
	if len(contracts) == 0 {
		return nil, nil
	}

	for _, contract := range contracts {
		id, ok := addresses[contract.Address]
		if !ok {
			return nil, errors.Errorf("can't find contract key: %s", contract.Address)
		}
		contract.Id = id
		if contract.TxId == nil {
			continue
		}

		txId, ok := txHashes[contract.Tx.Hash.String()]
		if !ok {
			return nil, errors.Errorf("can't find tx hash: %s", contract.Tx.Hash.String())
		}
		contract.TxId = &txId
	}

	err := tx.SaveContracts(ctx, contracts...)
	if err != nil {
		return nil, err
	}

	contractToId := make(map[string]uint64)
	for i := range contracts {
		contractToId[contracts[i].Address] = contracts[i].Id
	}

	return contractToId, err
}

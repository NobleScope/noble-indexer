package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
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
		contract.TxId = txHashes[contract.Tx.Hash.String()]
		contract.Id = addresses[contract.Address]
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

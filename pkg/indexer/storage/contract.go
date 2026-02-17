package storage

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
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

	for i := range contracts {
		id, ok := addresses[contracts[i].Address.Hash.String()]
		if !ok {
			return 0, errors.Errorf("can't find contract key: %s", contracts[i].Address.Hash.String())
		}
		contracts[i].Id = id

		if contracts[i].Deployer != nil {
			deployerId, ok := addresses[contracts[i].Deployer.Hash.String()]
			if !ok {
				return 0, errors.Errorf("can't find deployer key: %s", contracts[i].Deployer.Hash.String())
			}
			contracts[i].DeployerId = &deployerId
		}

		if contracts[i].Tx != nil {
			txId, ok := txHashes[contracts[i].Tx.Hash.String()]
			if !ok {
				return 0, errors.Errorf("can't find tx hash: %s", contracts[i].Tx.Hash.String())
			}
			contracts[i].TxId = &txId
		}
	}

	totalContracts, err := tx.SaveContracts(ctx, contracts...)
	if err != nil {
		return 0, err
	}

	return totalContracts, err
}

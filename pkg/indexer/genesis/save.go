package genesis

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/pkg/errors"
)

func (module *Module) save(ctx context.Context, data parsedData) error {
	start := time.Now()
	module.Log.Info().Msg("saving genesis block...")
	tx, err := postgres.BeginTransaction(ctx, module.storage.Transactable)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	if len(data.addresses) == 0 {
		return nil
	}

	var totalAccounts int64
	addresses := make([]*storage.Address, 0, len(data.addresses))
	addrToId := make(map[string]uint64)
	for key := range data.addresses {
		addresses = append(addresses, data.addresses[key])
	}

	totalAccounts, err = tx.SaveAddresses(ctx, addresses...)
	if err != nil {
		return tx.HandleError(ctx, err)
	}

	balances := make([]*storage.Balance, 0, len(data.addresses))
	for i := range addresses {
		addrToId[addresses[i].Hash.String()] = addresses[i].Id
		addresses[i].Balance.Id = addresses[i].Id
		balances = append(balances, addresses[i].Balance)
	}

	if err := tx.SaveBalances(ctx, balances...); err != nil {
		return tx.HandleError(ctx, err)
	}

	totalContracts := int64(0)
	if len(data.contracts) > 0 {
		entities := make([]*storage.Contract, 0, len(data.contracts))
		for key := range data.contracts {
			entities = append(entities, data.contracts[key])
		}

		for _, contract := range entities {
			id, ok := addrToId[contract.Address.Hash.String()]
			if !ok {
				return errors.Errorf("can't find contract key: %s", contract.Address.Hash.String())
			}
			contract.Id = id
		}

		totalContracts, err = tx.SaveContracts(ctx, entities...)
		if err != nil {
			return tx.HandleError(ctx, err)
		}
	}

	if err := tx.Add(ctx, &storage.State{
		Name:                   module.indexerName,
		ChainId:                data.chainId,
		LastHeight:             0,
		LastTime:               data.time,
		TotalTx:                0,
		TotalAccounts:          totalAccounts,
		TotalContracts:         totalContracts,
		TotalVerifiedContracts: 0,
		TotalTokens:            0,
	}); err != nil {
		return tx.HandleError(ctx, err)
	}

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}
	module.Log.Info().
		Int64("ms", time.Since(start).Milliseconds()).
		Msg("genesis saved")

	return nil
}

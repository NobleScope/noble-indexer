package genesis

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
)

func (module *Module) save(ctx context.Context, data parsedData) error {
	start := time.Now()
	module.Log.Info().Msg("saving genesis block...")
	tx, err := postgres.BeginTransaction(ctx, module.storage.Transactable)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	var totalAccounts int64
	if len(data.addresses) > 0 {
		entities := make([]*storage.Address, 0, len(data.addresses))
		for key := range data.addresses {
			entities = append(entities, data.addresses[key])
		}

		totalAccounts, err = tx.SaveAddresses(ctx, entities...)
		if err != nil {
			return tx.HandleError(ctx, err)
		}

		balances := make([]*storage.Balance, 0)
		for i := range entities {
			for _, b := range entities[i].Balance {
				b.Id = entities[i].Id
				balances = append(balances, b)
			}
		}

		if err := tx.SaveBalances(ctx, balances...); err != nil {
			return tx.HandleError(ctx, err)
		}
	}

	if len(data.contracts) > 0 {
		entities := make([]*storage.Contract, 0, len(data.contracts))
		for key := range data.contracts {
			entities = append(entities, data.contracts[key])
		}

		err = tx.SaveContracts(ctx, entities...)
		if err != nil {
			return tx.HandleError(ctx, err)
		}
	}

	if err := tx.Add(ctx, &storage.State{
		Name:          module.indexerName,
		ChainId:       data.chainId,
		LastHeight:    0,
		LastTime:      data.time,
		TotalTx:       0,
		TotalAccounts: totalAccounts,
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

package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
)

func saveTraces(
	ctx context.Context,
	tx storage.Transaction,
	traces []*storage.Trace,
	txHashes map[string]uint64,
	addresses map[string]uint64,
	contracts map[string]uint64,
) error {
	if len(traces) == 0 {
		return nil
	}
	for i := range traces {
		traces[i].TxId = txHashes[traces[i].Tx.Hash.String()]
		traces[i].From = addresses[traces[i].FromAddress.Address]

		if traces[i].ToAddress != nil {
			id := addresses[traces[i].ToAddress.Address]
			traces[i].To = &id
		}

		if traces[i].Tx.Contract != nil {
			id := contracts[traces[i].Tx.Contract.Address]
			traces[i].ContractId = &id
		}
	}

	err := tx.SaveTraces(ctx, traces...)
	return err
}

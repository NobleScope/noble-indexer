package storage

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/pkg/errors"
)

func saveTraces(
	ctx context.Context,
	tx storage.Transaction,
	traces []*storage.Trace,
	txHashes map[string]uint64,
	addresses map[string]uint64,
) error {
	if len(traces) == 0 {
		return nil
	}
	for i := range traces {
		if traces[i].Tx != nil {
			id, ok := txHashes[traces[i].Tx.Hash.String()]
			if !ok {
				traces[i].TxId = nil
			} else {
				traces[i].TxId = &id
			}
		}

		if traces[i].FromAddress != nil {
			id, ok := addresses[traces[i].FromAddress.String()]
			if !ok {
				return errors.Errorf("can't find address key: %s", traces[i].FromAddress.String())
			}
			traces[i].From = &id
		}

		if traces[i].ToAddress != nil {
			id, ok := addresses[traces[i].ToAddress.String()]
			if !ok {
				return errors.Errorf("can't find address key: %s", traces[i].ToAddress.String())
			}
			traces[i].To = &id
		}

		if traces[i].Contract != nil {
			id, ok := addresses[traces[i].Contract.Address.String()]
			if !ok {
				return errors.Errorf("can't find address key: %s", traces[i].Contract.Address.String())
			}
			traces[i].ContractId = &id
		}
	}

	err := tx.SaveTraces(ctx, traces...)
	return err
}

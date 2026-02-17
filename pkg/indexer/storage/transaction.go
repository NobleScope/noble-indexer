package storage

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
)

func saveTransactions(
	ctx context.Context,
	tx storage.Transaction,
	txs []*storage.Tx,
	addresses map[string]uint64,
) error {
	if len(txs) == 0 {
		return nil
	}

	for _, t := range txs {
		t.FromAddressId = addresses[t.FromAddress.String()]

		if t.ToAddress != nil {
			id := addresses[t.ToAddress.String()]
			t.ToAddressId = &id
		}
	}

	err := tx.SaveTransactions(ctx, txs...)

	return err
}

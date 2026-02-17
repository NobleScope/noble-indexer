package storage

import (
	"context"

	"github.com/pkg/errors"

	"github.com/NobleScope/noble-indexer/internal/pool"
	"github.com/NobleScope/noble-indexer/internal/storage"
)

var logsPool = pool.New(func() []*storage.Log {
	return make([]*storage.Log, 0, 1024)
})

func saveLogs(
	ctx context.Context,
	tx storage.Transaction,
	transactions []*storage.Tx,
	addresses map[string]uint64,
) error {
	logs := logsPool.Get()
	defer func() {
		for i := range logs {
			logs[i] = nil
		}
		logs = logs[:0]
		logsPool.Put(logs)
	}()

	for i := range transactions {
		for j := range transactions[i].Logs {
			transactions[i].Logs[j].TxId = transactions[i].Id

			id, ok := addresses[transactions[i].Logs[j].Address.String()]
			if !ok {
				return errors.Errorf("can't find log address key: %s", transactions[i].Logs[j].Address.String())
			}
			transactions[i].Logs[j].AddressId = id
		}

		logs = append(logs, transactions[i].Logs...)
	}

	if len(logs) == 0 {
		return nil
	}

	err := tx.SaveLogs(ctx, logs...)

	return err
}

package rollback

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/currency"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/shopspring/decimal"
)

func (module *Module) rollbackBalances(
	ctx context.Context,
	tx storage.Transaction,
	block storage.Block,
	deletedTxs []storage.Tx,
	deletedTraces []storage.Trace,
	deletedAddresses []storage.Address,
) error {
	var (
		ids     = make([]uint64, len(deletedAddresses))
		deleted = make(map[uint64]struct{}, len(deletedAddresses))
	)
	for i := range deletedAddresses {
		ids[i] = deletedAddresses[i].Id
		deleted[deletedAddresses[i].Id] = struct{}{}
	}

	if err := tx.DeleteBalances(ctx, ids); err != nil {
		return err
	}

	if len(deletedTxs) == 0 {
		return nil
	}

	updates, err := getBalanceUpdates(block, deleted, deletedTxs, deletedTraces)
	if err != nil {
		return err
	}

	err = tx.SaveBalances(ctx, updates...)
	return err
}

func getBalanceUpdates(
	block storage.Block,
	deletedAddressIds map[uint64]struct{},
	deletedTxs []storage.Tx,
	deletedTraces []storage.Trace,
) ([]*storage.Balance, error) {
	updates := make(map[uint64]decimal.Decimal)
	for _, t := range deletedTxs {
		if t.Status == types.TxStatusRevert {
			continue
		}

		if _, ok := deletedAddressIds[t.FromAddressId]; !ok {
			updates[t.FromAddressId] = t.Amount.Add(t.Fee)
		}

		if t.ToAddressId != nil {
			if _, ok := deletedAddressIds[*t.ToAddressId]; !ok {
				updates[*t.ToAddressId] = t.Amount.Neg()
			}
		}

		if _, ok := deletedAddressIds[block.MinerId]; !ok {
			burnedFee := t.CumulativeGasUsed.Mul(decimal.NewFromUint64(block.BaseFeePerGas))
			fee := t.Fee.Sub(burnedFee)
			updates[block.MinerId] = fee.Neg()
		}
	}

	for _, trace := range deletedTraces {
		if len(trace.TraceAddress) == 0 || trace.Amount == nil || trace.Amount.IsZero() {
			continue
		}

		if _, ok := deletedAddressIds[trace.From]; !ok {
			updates[trace.From] = *trace.Amount
		}
		if trace.ToAddress != nil {
			if _, ok := deletedAddressIds[*trace.To]; !ok {
				updates[*trace.To] = trace.Amount.Neg()
			}
		}
	}

	result := make([]*storage.Balance, 0, len(updates))
	for i := range updates {
		balance := &storage.Balance{
			Id:       i,
			Currency: currency.DefaultCurrency,
			Value:    updates[i],
		}
		result = append(result, balance)
	}
	// TODO: update LastHeight for addresses
	return result, nil
}

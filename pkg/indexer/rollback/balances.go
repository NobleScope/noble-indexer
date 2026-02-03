package rollback

import (
	"context"

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
	deletedTransfers []storage.Transfer,
	deletedTokens []storage.Token,
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
	if err != nil {
		return err
	}

	tokenUpdates, err := getTokenBalanceUpdates(deletedTransfers)
	if err != nil {
		return err
	}

	tbs, err := tx.SaveTokenBalances(ctx, tokenUpdates...)
	if err != nil {
		return err
	}

	var zeroTokenBalances = make([]*storage.TokenBalance, 0)
	for i := range tbs {
		if tbs[i].Balance.IsZero() {
			zeroTokenBalances = append(zeroTokenBalances, &tbs[i])
		}
	}

	var (
		tokenIds    = make([]string, len(deletedTokens))
		contractIds = make([]uint64, len(deletedTokens))
	)

	for i, t := range deletedTokens {
		tokenIds[i] = t.TokenID.String()
		contractIds[i] = t.ContractId
	}

	err = tx.DeleteTokenBalances(ctx, tokenIds, contractIds, zeroTokenBalances)

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
		if trace.FromAddress != nil {
			if _, ok := deletedAddressIds[*trace.From]; !ok {
				updates[*trace.From] = *trace.Amount
			}
		}
		if trace.ToAddress != nil {
			if _, ok := deletedAddressIds[*trace.To]; !ok {
				updates[*trace.To] = trace.Amount.Neg()
			}
		}
	}

	result := make([]*storage.Balance, len(updates))
	for i := range updates {
		result[i] = &storage.Balance{
			Id:    i,
			Value: updates[i],
		}
	}
	// TODO: update LastHeight for addresses
	return result, nil
}

func getTokenBalanceUpdates(
	deletedTransfers []storage.Transfer,
) ([]*storage.TokenBalance, error) {
	type key struct {
		TokenID    string
		ContractID uint64
		AddressID  uint64
	}

	agg := make(map[key]*storage.TokenBalance)

	add := func(tokenID decimal.Decimal, contractID uint64, addrID uint64, amount decimal.Decimal) {
		k := key{tokenID.String(), contractID, addrID}

		if tb, ok := agg[k]; ok {
			tb.Balance = tb.Balance.Add(amount)
		} else {
			agg[k] = &storage.TokenBalance{
				TokenID:    tokenID,
				ContractID: contractID,
				AddressID:  addrID,
				Balance:    amount,
			}
		}
	}

	for _, t := range deletedTransfers {
		switch t.Type {

		case types.Mint:
			add(t.TokenID, t.ContractId, *t.ToAddressId, t.Amount.Neg())

		case types.Burn:
			add(t.TokenID, t.ContractId, *t.FromAddressId, t.Amount)

		case types.Transfer:
			add(t.TokenID, t.ContractId, *t.FromAddressId, t.Amount)
			add(t.TokenID, t.ContractId, *t.ToAddressId, t.Amount.Neg())
		}
	}

	result := make([]*storage.TokenBalance, 0, len(agg))
	for _, v := range agg {
		result = append(result, v)
	}

	return result, nil
}

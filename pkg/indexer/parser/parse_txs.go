package parser

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/baking-bad/noble-indexer/pkg/indexer/enum"
	"github.com/shopspring/decimal"
)

func (p *Module) parseTxs(context *dCtx.Context) error {
	for _, tx := range context.Block.Txs {
		tx.TracesCount = len(context.GetTracesByTxHash(tx.Hash))

		if tx.Status == types.TxStatusRevert {
			continue
		}

		totalAmount := tx.Amount.Add(tx.Fee)
		updateAddressBalance(context, tx.FromAddress.String(), enum.Sub, totalAmount)

		if tx.ToAddress != nil {
			updateAddressBalance(context, tx.ToAddress.String(), enum.Add, tx.Amount)
		}

		burnedFee := tx.CumulativeGasUsed.Mul(decimal.NewFromUint64(context.Block.BaseFeePerGas))
		fee := tx.Fee.Sub(burnedFee)
		updateAddressBalance(context, context.Block.Miner.String(), enum.Add, fee)
	}

	for _, trace := range context.Block.Traces {
		if trace.Amount == nil || trace.Amount.IsZero() {
			continue
		}

		isCreate := trace.Type == types.Create || trace.Type == types.Create2

		if len(trace.TraceAddress) == 0 {
			if isCreate && trace.Contract != nil {
				updateAddressBalance(context, trace.Contract.Address.String(), enum.Add, *trace.Amount)
			}
			continue
		}

		if trace.FromAddress != nil {
			updateAddressBalance(context, trace.FromAddress.String(), enum.Sub, *trace.Amount)
		}

		if isCreate && trace.ToAddress == nil && trace.Contract != nil {
			updateAddressBalance(context, trace.Contract.Address.String(), enum.Add, *trace.Amount)
		} else if trace.ToAddress != nil {
			updateAddressBalance(context, trace.ToAddress.String(), enum.Add, *trace.Amount)
		}
	}

	return nil
}

func updateAddressBalance(ctx *dCtx.Context, addressKey string, op enum.BalanceOp, amount decimal.Decimal) {
	addr, ok := ctx.Addresses.Get(addressKey)
	if !ok {
		return
	}
	updateBalances(addr, op, amount)
}

func updateBalances(address *storage.Address, op enum.BalanceOp, amount decimal.Decimal) {
	if address == nil {
		return
	}

	if address.Balance != nil {
		switch op {
		case enum.Add:
			address.Balance.Value = address.Balance.Value.Add(amount)
		case enum.Sub:
			address.Balance.Value = address.Balance.Value.Sub(amount)
		}
		return
	}

	initial := decimal.Zero
	switch op {
	case enum.Add:
		initial = amount
	case enum.Sub:
		initial = decimal.Zero.Sub(amount)
	}

	address.Balance = &storage.Balance{
		Value: initial,
	}
}

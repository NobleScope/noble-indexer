package parser

import (
	"github.com/baking-bad/noble-indexer/internal/currency"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/baking-bad/noble-indexer/pkg/indexer/enum"
	"github.com/shopspring/decimal"
)

func (p *Module) parseTxs(context *dCtx.Context) error {
	// update balances on tx
	for _, tx := range context.Block.Txs {
		if tx.Status == types.TxStatusRevert {
			continue
		}

		totalAmount := tx.Amount.Add(tx.Fee)
		updateBalances(&tx.FromAddress, enum.Sub, totalAmount)
		updateBalances(tx.ToAddress, enum.Add, tx.Amount)

		minerAddress := &storage.Address{
			Address:    context.Block.MinerHash.String(),
			Height:     context.Block.Height,
			LastHeight: context.Block.Height,
		}
		context.AddAddress(minerAddress)
		burnedFee := tx.CumulativeGasUsed.Mul(decimal.NewFromUint64(context.Block.BaseFeePerGas))
		fee := tx.Fee.Sub(burnedFee)
		updateBalances(minerAddress, enum.Add, fee)
	}

	// update balances on traces
	for _, trace := range context.Block.Traces {
		if len(trace.TraceAddress) == 0 || trace.Amount == nil || trace.Amount.IsZero() {
			continue
		}

		updateBalances(&trace.FromAddress, enum.Sub, *trace.Amount)
		updateBalances(trace.ToAddress, enum.Add, *trace.Amount)
	}

	return nil
}

func updateBalances(address *storage.Address, op enum.BalanceOp, amount decimal.Decimal) {
	if address == nil {
		return
	}

	if len(address.Balance) > 0 {
		for _, b := range address.Balance {
			if b.Currency == currency.DefaultCurrency {
				switch op {
				case enum.Add:
					b.Value = b.Value.Add(amount)
				case enum.Sub:
					b.Value = b.Value.Sub(amount)
				}
			}
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

	address.Balance = []*storage.Balance{
		{
			Currency: currency.DefaultCurrency,
			Value:    initial,
		},
	}
}

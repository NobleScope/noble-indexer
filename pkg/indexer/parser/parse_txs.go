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
		updateBalances(&tx.FromAddress, enum.Sub, totalAmount)
		updateBalances(tx.ToAddress, enum.Add, tx.Amount)

		burnedFee := tx.CumulativeGasUsed.Mul(decimal.NewFromUint64(context.Block.BaseFeePerGas))
		fee := tx.Fee.Sub(burnedFee)
		updateBalances(&context.Block.Miner, enum.Add, fee)
	}

	for _, trace := range context.Block.Traces {
		if len(trace.TraceAddress) == 0 || trace.Amount == nil || trace.Amount.IsZero() {
			continue
		}

		updateBalances(trace.FromAddress, enum.Sub, *trace.Amount)
		updateBalances(trace.ToAddress, enum.Add, *trace.Amount)
	}

	return nil
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

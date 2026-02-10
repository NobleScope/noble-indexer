package parser

import (
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
	dCtx "github.com/NobleScope/noble-indexer/pkg/indexer/decode/context"
	"github.com/NobleScope/noble-indexer/pkg/indexer/enum"
	"github.com/shopspring/decimal"
)

func (p *Module) parseTxs(context *dCtx.Context) error {
	for i := range context.Block.Txs {
		traces := context.GetTracesByTxHash(context.Block.Txs[i].Hash)
		context.Block.Txs[i].TracesCount = len(traces)

		if context.Block.Txs[i].Status != types.TxStatusSuccess {
			continue
		}

		totalAmount := context.Block.Txs[i].Amount.Add(context.Block.Txs[i].Fee)
		updateAddressBalance(context, context.Block.Txs[i].FromAddress.String(), enum.Sub, totalAmount)

		if context.Block.Txs[i].ToAddress != nil {
			updateAddressBalance(context, context.Block.Txs[i].ToAddress.String(), enum.Add, context.Block.Txs[i].Amount)
		}

		burnedFee := context.Block.Txs[i].GasUsed.Mul(decimal.NewFromUint64(context.Block.BaseFeePerGas))
		fee := context.Block.Txs[i].Fee.Sub(burnedFee)
		updateAddressBalance(context, context.Block.Miner.String(), enum.Add, fee)

		for j := range traces {
			if traces[j].Amount == nil || traces[j].Amount.IsZero() {
				continue
			}

			isCreate := traces[j].Type == types.Create || traces[j].Type == types.Create2

			if len(traces[j].TraceAddress) == 0 {
				if isCreate && traces[j].Contract != nil {
					updateAddressBalance(context, traces[j].Contract.Address.String(), enum.Add, *traces[j].Amount)
				}
				continue
			}

			if traces[j].FromAddress != nil {
				updateAddressBalance(context, traces[j].FromAddress.String(), enum.Sub, *traces[j].Amount)
			}

			if isCreate && traces[j].ToAddress == nil && traces[j].Contract != nil {
				updateAddressBalance(context, traces[j].Contract.Address.String(), enum.Add, *traces[j].Amount)
			} else if traces[j].ToAddress != nil {
				updateAddressBalance(context, traces[j].ToAddress.String(), enum.Add, *traces[j].Amount)
			}
		}
	}

	for i := range context.Block.Traces {
		if context.Block.Traces[i].Type != types.Reward {
			continue
		}
		if context.Block.Traces[i].Amount == nil || context.Block.Traces[i].Amount.IsZero() {
			continue
		}
		if context.Block.Traces[i].ToAddress != nil {
			updateAddressBalance(context, context.Block.Traces[i].ToAddress.String(), enum.Add, *context.Block.Traces[i].Amount)
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

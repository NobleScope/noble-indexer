package parser

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/currency"
	"github.com/baking-bad/noble-indexer/internal/storage"
	storageType "github.com/baking-bad/noble-indexer/internal/storage/types"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/shopspring/decimal"
)

func (p *Module) parse(b types.BlockData) error {
	start := time.Now()
	decodeCtx := dCtx.NewContext()

	height, err := b.Number.Uint64()
	if err != nil {
		return err
	}
	p.Log.Info().
		Uint64("height", height).
		Msg("parsing block...")

	p.Log.Info().
		Uint64("height", height).
		Int64("ms", time.Since(start).Milliseconds()).
		Msg("block parsed")

	block := b.Block
	blockTime, err := block.Timestamp.Time()
	if err != nil {
		return err
	}
	gasLimit, err := block.GasLimit.Decimal()
	if err != nil {
		return err
	}
	gasUsed, err := block.GasUsed.Decimal()
	if err != nil {
		return err
	}
	feePerGas, err := block.BaseFeePerGas.Uint64()
	if err != nil {
		return err
	}

	miner := storage.Address{
		Address:    block.Miner.String(),
		Height:     types.Level(height),
		LastHeight: types.Level(height),
		Balance: []*storage.Balance{
			{
				Currency: currency.DefaultCurrency,
				Value:    decimal.Zero,
			},
		},
	}
	decodeCtx.AddAddress(&miner)

	decodeCtx.Block = &storage.Block{
		Time:                 blockTime,
		Height:               types.Level(height),
		GasLimit:             gasLimit,
		GasUsed:              gasUsed,
		BaseFeePerGas:        feePerGas,
		Miner:                miner,
		DifficultyHash:       b.Difficulty,
		ExtraDataHash:        b.ExtraData,
		Hash:                 b.Hash,
		LogsBloomHash:        b.LogsBloom,
		MixHash:              b.MixHash,
		NonceHash:            b.Nonce,
		ParentHashHash:       b.ParentHash,
		ReceiptsRootHash:     b.ReceiptsRoot,
		Sha3UnclesHash:       b.Sha3Uncles,
		StateRootHash:        b.StateRoot,
		SizeHash:             b.Size,
		TotalDifficultyHash:  b.TotalDifficulty,
		TransactionsRootHash: b.TransactionsRoot,
		Txs:                  make([]*storage.Tx, len(b.Transactions)),
		Traces:               make([]*storage.Trace, len(b.Traces)),
		Stats: &storage.BlockStats{
			Height:  types.Level(height),
			Time:    blockTime,
			TxCount: int64(len(b.Transactions)),
		},
	}

	for i, tx := range b.Transactions {
		gas, err := tx.Gas.Decimal()
		if err != nil {
			return err
		}
		gasPrice, err := tx.GasPrice.Decimal()
		if err != nil {
			return err
		}
		nonce, err := tx.Nonce.Int64()
		if err != nil {
			return err
		}
		index, err := tx.TransactionIndex.Int64()
		if err != nil {
			return err
		}

		var txType storageType.TxType
		switch tx.Type.String() {
		case "0x00":
			txType = storageType.TxTypeLegacy
		case "0x02":
			txType = storageType.TxTypeDynamicFee
		default:
			txType = storageType.TxTypeUnknown
		}

		cumulativeGasUsed, err := b.Receipts[i].CumulativeGasUsed.Decimal()
		if err != nil {
			return err
		}
		effectiveGasPrice, err := b.Receipts[i].EffectiveGasPrice.Decimal()
		if err != nil {
			return err
		}
		txGasUsed, err := b.Receipts[i].GasUsed.Decimal()
		if err != nil {
			return err
		}
		amount, err := tx.Value.Decimal()
		if err != nil {
			return err
		}

		fee := cumulativeGasUsed.Mul(effectiveGasPrice)

		var status storageType.TxStatus
		switch b.Receipts[i].Status.String() {
		case "0x01":
			status = storageType.TxStatusSuccess
		case "0x00":
			status = storageType.TxStatusRevert
		default:
			status = storageType.TxStatusRevert
		}

		decodeCtx.Block.Txs[i] = &storage.Tx{
			Height:   types.Level(height),
			Time:     blockTime,
			Gas:      gas,
			GasPrice: gasPrice,
			Hash:     tx.Hash,
			Nonce:    nonce,
			Index:    index,
			Amount:   amount,
			Type:     txType,
			Input:    tx.Input,

			CumulativeGasUsed: cumulativeGasUsed,
			EffectiveGasPrice: effectiveGasPrice,
			Fee:               fee,
			GasUsed:           txGasUsed,
			Status:            status,
			LogsBloom:         b.Receipts[i].LogsBloom,
			Logs:              make([]storage.Log, len(b.Receipts[i].Logs)),
			LogsCount:         len(b.Receipts[i].Logs),
		}

		decodeCtx.Block.Txs[i].FromAddress = storage.Address{
			Address:      b.Receipts[i].From.String(),
			Height:       decodeCtx.Block.Height,
			LastHeight:   decodeCtx.Block.Height,
			Interactions: 1,
			TxsCount:     1,
			Balance: []*storage.Balance{
				{
					Currency: currency.DefaultCurrency,
					Value:    decimal.Zero,
				},
			},
		}

		if b.Receipts[i].ContractAddress != nil {
			contractAddress := storage.Address{
				Address:    b.Receipts[i].ContractAddress.String(),
				Height:     decodeCtx.Block.Height,
				LastHeight: decodeCtx.Block.Height,
				IsContract: true,
			}

			contract := &storage.Contract{
				Address: b.Receipts[i].ContractAddress.String(),
				Code:    decodeCtx.Block.Txs[i].Input,
				Tx: &storage.Tx{
					Hash: b.Receipts[i].TransactionHash,
				},
			}

			err = ParseEvmContractMetadata(contract)
			if err != nil {
				return err
			}

			decodeCtx.Block.Txs[i].Contract = contract
			decodeCtx.Block.Txs[i].FromAddress.ContractsCount = 1

			decodeCtx.AddAddress(&contractAddress)
			decodeCtx.AddContract(decodeCtx.Block.Txs[i].Contract)
		}

		decodeCtx.AddAddress(&decodeCtx.Block.Txs[i].FromAddress)

		if b.Transactions[i].To != nil {
			decodeCtx.Block.Txs[i].ToAddress = &storage.Address{
				Address:      b.Receipts[i].To.String(),
				Height:       decodeCtx.Block.Height,
				LastHeight:   decodeCtx.Block.Height,
				Interactions: 1,
				Balance: []*storage.Balance{
					{
						Currency: currency.DefaultCurrency,
						Value:    decimal.Zero,
					},
				},
			}

			if b.Transactions[i].From.String() == b.Transactions[i].To.String() {
				decodeCtx.Block.Txs[i].ToAddress.Interactions = 0
			}

			decodeCtx.AddAddress(decodeCtx.Block.Txs[i].ToAddress)
		}

		for j, log := range b.Receipts[i].Logs {
			logIndex, err := log.LogIndex.Int64()
			if err != nil {
				return err
			}

			var name string
			if len(log.Topics) > 0 {
				name = log.Topics[0].String()
			}

			decodeCtx.Block.Txs[i].Logs[j] = storage.Log{
				Height:  types.Level(height),
				Time:    blockTime,
				Index:   logIndex,
				Name:    name,
				Data:    log.Data,
				Topics:  log.Topics,
				Address: log.Address,
				Removed: log.Removed,
			}
		}
	}

	for i, trace := range b.Traces {
		gl, err := trace.Action.Gas.Decimal()
		if err != nil {
			return err
		}

		gu, err := trace.Result.GasUsed.Decimal()
		if err != nil {
			return err
		}

		value, err := trace.Action.Value.Decimal()
		if err != nil {
			return err
		}

		typ, err := storageType.ParseTraceType(trace.Type)
		if err != nil {
			return err
		}

		newTrace := &storage.Trace{
			Height: types.Level(height),
			Time:   blockTime,

			FromAddress: storage.Address{
				Address:      trace.Action.From.String(),
				Height:       decodeCtx.Block.Height,
				LastHeight:   decodeCtx.Block.Height,
				Interactions: 1,
				Balance: []*storage.Balance{
					{
						Currency: currency.DefaultCurrency,
						Value:    decimal.Zero,
					},
				},
			},
			Tx: storage.Tx{
				Hash: trace.TxHash,
			},

			TraceAddress:   trace.TraceAddress,
			TxPosition:     trace.TxPosition,
			GasLimit:       gl,
			Amount:         &value,
			Type:           typ,
			InitHash:       trace.Action.Init,
			CreationMethod: trace.Action.CreationMethod,

			GasUsed: gu,
			Code:    trace.Result.Code,

			Subtraces: trace.Subtraces,
		}

		if trace.Result.Address != nil {
			newTrace.Tx.Contract = &storage.Contract{
				Address: trace.Result.Address.String(),
			}
		}

		decodeCtx.AddAddress(&newTrace.FromAddress)

		if trace.Action.To != nil {
			newTrace.ToAddress = &storage.Address{
				Address:      trace.Action.To.String(),
				Height:       decodeCtx.Block.Height,
				LastHeight:   decodeCtx.Block.Height,
				Interactions: 1,
				Balance: []*storage.Balance{
					{
						Currency: currency.DefaultCurrency,
						Value:    decimal.Zero,
					},
				},
			}
			if trace.Action.From.String() != trace.Action.To.String() {
				newTrace.ToAddress.Interactions = 0
			}

			decodeCtx.AddAddress(newTrace.ToAddress)
		}

		if trace.Action.Input != nil {
			newTrace.Input = trace.Action.Input.Bytes()
		}

		if trace.Result.Output != nil {
			newTrace.Output = trace.Result.Output.Bytes()
		}

		decodeCtx.Block.Traces[i] = newTrace
		decodeCtx.AddTrace(newTrace)
	}

	err = p.parseTxs(decodeCtx)
	if err != nil {
		return err
	}

	err = p.parseTransfers(decodeCtx)
	if err != nil {
		return err
	}

	output := p.MustOutput(OutputName)
	output.Push(decodeCtx)

	return nil
}

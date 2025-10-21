package parser

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	storageType "github.com/baking-bad/noble-indexer/internal/storage/types"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/baking-bad/noble-indexer/pkg/types"
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

	blockData := b.Block
	blockTime, err := blockData.Timestamp.Time()
	if err != nil {
		return err
	}
	gasLimit, err := blockData.GasLimit.Uint64()
	if err != nil {
		return err
	}
	gasUsed, err := blockData.GasUsed.Uint64()
	if err != nil {
		return err
	}

	decodeCtx.Block = &storage.Block{
		Time:                 blockTime,
		Height:               types.Level(height),
		GasLimit:             gasLimit,
		GasUsed:              gasUsed,
		DifficultyHash:       b.Difficulty,
		ExtraDataHash:        b.ExtraData,
		Hash:                 b.Hash,
		LogsBloomHash:        b.LogsBloom,
		MinerHash:            b.Miner,
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
	}

	for i, tx := range b.Transactions {
		gas, err := tx.Gas.Int64()
		if err != nil {
			return err
		}
		gasPrice, err := tx.GasPrice.Int64()
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
		value, err := tx.Value.Int64()
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

		cumulativeGasUsed, err := b.Receipts[i].CumulativeGasUsed.Int64()
		if err != nil {
			return err
		}
		effectiveGasPrice, err := b.Receipts[i].EffectiveGasPrice.Int64()
		if err != nil {
			return err
		}
		fee, err := b.Receipts[i].L1Fee.Decimal()
		if err != nil {
			return err
		}
		txGasUsed, err := b.Receipts[i].GasUsed.Int64()
		if err != nil {
			return err
		}
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
			Value:    value,
			Type:     txType,
			Input:    tx.Input,

			CumulativeGasUsed: cumulativeGasUsed,
			EffectiveGasPrice: effectiveGasPrice,
			Fee:               fee,
			GasUsed:           txGasUsed,
			Status:            status,
			LogsBloom:         b.Receipts[i].LogsBloom,
			Logs:              make([]storage.Log, len(b.Receipts[i].Logs)),
		}

		if b.Receipts[i].ContractAddress != nil {
			contractAddress := storage.Address{
				Address:    b.Receipts[i].ContractAddress.String(),
				Height:     decodeCtx.Block.Height,
				LastHeight: decodeCtx.Block.Height,
				IsContract: true,
			}

			decodeCtx.Block.Txs[i].Contract = &storage.Contract{
				Address: b.Receipts[i].ContractAddress.String(),
				Code:    decodeCtx.Block.Txs[i].Input,
				Tx: &storage.Tx{
					Hash: b.Receipts[i].TransactionHash,
				},
			}

			decodeCtx.AddAddress(&contractAddress)
			decodeCtx.AddContract(decodeCtx.Block.Txs[i].Contract)
		}

		decodeCtx.Block.Txs[i].FromAddress = storage.Address{
			Address:    b.Receipts[i].From.String(),
			Height:     decodeCtx.Block.Height,
			LastHeight: decodeCtx.Block.Height,
		}

		decodeCtx.AddAddress(&decodeCtx.Block.Txs[i].FromAddress)

		if b.Transactions[i].To != nil {
			decodeCtx.Block.Txs[i].ToAddress = &storage.Address{
				Address:    b.Receipts[i].To.String(),
				Height:     decodeCtx.Block.Height,
				LastHeight: decodeCtx.Block.Height,
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
				Removed: log.Removed,
			}
		}

	}

	for i, trace := range b.Traces {
		gasLimit, err = trace.Action.Gas.Uint64()
		if err != nil {
			return err
		}

		gasUsed, err = trace.Result.GasUsed.Uint64()
		if err != nil {
			return err
		}

		value, err := trace.Action.Value.Uint64()
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
				Address:    trace.Action.From.String(),
				Height:     decodeCtx.Block.Height,
				LastHeight: decodeCtx.Block.Height,
			},
			Tx: storage.Tx{
				Hash: trace.TxHash,
			},

			GasLimit:       gasLimit,
			Value:          &value,
			Type:           typ,
			InitHash:       trace.Action.Init,
			CreationMethod: trace.Action.CreationMethod,

			GasUsed: gasUsed,
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
				Address:    trace.Action.To.String(),
				Height:     decodeCtx.Block.Height,
				LastHeight: decodeCtx.Block.Height,
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

	output := p.MustOutput(OutputName)
	output.Push(decodeCtx)

	return nil
}

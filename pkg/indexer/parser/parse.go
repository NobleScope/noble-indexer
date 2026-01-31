package parser

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	storageType "github.com/baking-bad/noble-indexer/internal/storage/types"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/pkg/errors"
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
		Hash:        block.Miner,
		FirstHeight: types.Level(height),
		LastHeight:  types.Level(height),
		Balance:     storage.EmptyBalance(),
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
		typ, err := tx.Type.Int64()
		if err != nil {
			return err
		}
		var txType storageType.TxType
		switch typ {
		case 0:
			txType = storageType.TxTypeLegacy
		case 2:
			txType = storageType.TxTypeDynamicFee
		case 3:
			txType = storageType.TxTypeBlob
		case 4:
			txType = storageType.TxTypeSetCode
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
		fee := cumulativeGasUsed.Mul(effectiveGasPrice)
		txGasUsed, err := b.Receipts[i].GasUsed.Decimal()
		if err != nil {
			return err
		}
		amount, err := tx.Value.Decimal()
		if err != nil {
			return err
		}
		txStatus, err := b.Receipts[i].Status.Int64()
		if err != nil {
			return err
		}

		var status storageType.TxStatus
		switch txStatus {
		case 1:
			status = storageType.TxStatusSuccess
		case 0:
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
			Logs:              make([]*storage.Log, len(b.Receipts[i].Logs)),
			LogsCount:         len(b.Receipts[i].Logs),
		}

		decodeCtx.Block.Txs[i].FromAddress = storage.Address{
			Hash:         b.Receipts[i].From,
			FirstHeight:  decodeCtx.Block.Height,
			LastHeight:   decodeCtx.Block.Height,
			Interactions: 1,
			TxsCount:     1,
			Balance:      storage.EmptyBalance(),
		}

		decodeCtx.AddAddress(&decodeCtx.Block.Txs[i].FromAddress)

		if b.Transactions[i].To != nil {
			decodeCtx.Block.Txs[i].ToAddress = &storage.Address{
				Hash:         b.Receipts[i].To,
				FirstHeight:  decodeCtx.Block.Height,
				LastHeight:   decodeCtx.Block.Height,
				Interactions: 1,
				Balance:      storage.EmptyBalance(),
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

			decodeCtx.Block.Txs[i].Logs[j] = &storage.Log{
				Height:  types.Level(height),
				Time:    blockTime,
				Index:   logIndex,
				Name:    name,
				Data:    log.Data,
				Topics:  log.Topics,
				Removed: log.Removed,
			}

			decodeCtx.Block.Txs[i].Logs[j].Address = storage.Address{
				Hash:         b.Receipts[i].Logs[j].Address,
				FirstHeight:  decodeCtx.Block.Height,
				LastHeight:   decodeCtx.Block.Height,
				Interactions: 1,
				Balance:      storage.EmptyBalance(),
			}

			decodeCtx.AddAddress(&decodeCtx.Block.Txs[i].Logs[j].Address)
		}

		p.parseEIP1967Proxy(decodeCtx, decodeCtx.Block.Txs[i].Logs)
	}

	for i, trace := range b.Traces {
		typ, err := storageType.ParseTraceType(trace.Type)
		if err != nil {
			return err
		}

		var value decimal.Decimal
		if trace.Action.Value != nil {
			value, err = trace.Action.Value.Decimal()
			if err != nil {
				return err
			}
		}

		newTrace := &storage.Trace{
			Height:         types.Level(height),
			Time:           blockTime,
			Amount:         &value,
			TraceAddress:   trace.TraceAddress,
			Type:           typ,
			Subtraces:      trace.Subtraces,
			InitHash:       trace.Action.Init,
			CreationMethod: trace.Action.CreationMethod,
		}

		if typ == storageType.Reward {
			newTrace.ToAddress = &storage.Address{
				Hash:        *trace.Action.Author,
				FirstHeight: decodeCtx.Block.Height,
				LastHeight:  decodeCtx.Block.Height,
				Balance:     storage.EmptyBalance(),
			}
			newTrace.Tx = &storage.Tx{
				Hash: nil,
			}

			decodeCtx.AddAddress(newTrace.ToAddress)

			decodeCtx.Block.Traces[i] = newTrace
			decodeCtx.AddTrace(newTrace)
			continue
		}

		var (
			gl, gu decimal.Decimal
		)

		if trace.Action.Gas != nil {
			gl, err = trace.Action.Gas.Decimal()
			if err != nil {
				return err
			}

			newTrace.GasLimit = gl
		}

		if trace.Result.GasUsed != nil {
			gu, err = trace.Result.GasUsed.Decimal()
			if err != nil {
				return err
			}

			newTrace.GasUsed = gu
		}

		if trace.TxPosition != nil {
			newTrace.TxPosition = trace.TxPosition
		}

		if trace.TxHash != nil {
			newTrace.Tx = &storage.Tx{
				Hash: *trace.TxHash,
			}
		}

		if trace.Action.From != nil {
			newTrace.FromAddress = &storage.Address{
				Hash:         *trace.Action.From,
				FirstHeight:  decodeCtx.Block.Height,
				LastHeight:   decodeCtx.Block.Height,
				Interactions: 1,
				Balance:      storage.EmptyBalance(),
			}

			decodeCtx.AddAddress(newTrace.FromAddress)
		}

		if trace.Action.To != nil {
			newTrace.ToAddress = &storage.Address{
				Hash:         *trace.Action.To,
				FirstHeight:  decodeCtx.Block.Height,
				LastHeight:   decodeCtx.Block.Height,
				Interactions: 1,
				Balance:      storage.EmptyBalance(),
			}
			if trace.Action.From != nil && trace.Action.From.String() != trace.Action.To.String() {
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

		if trace.Result.Address != nil && trace.Result.Code != nil && len(*trace.Result.Code) > 0 {
			if trace.TxPosition == nil {
				return errors.New("trace.TxPosition is nil for contract creation")
			}
			if *trace.TxPosition >= uint64(len(b.Transactions)) {
				return errors.Errorf(
					"TxPosition %d out of range, transactions count: %d",
					*trace.TxPosition,
					len(b.Transactions),
				)
			}

			deployerAddress := storage.Address{
				Hash:           b.Transactions[*trace.TxPosition].From,
				LastHeight:     decodeCtx.Block.Height,
				Balance:        storage.EmptyBalance(),
				ContractsCount: 1,
			}
			contractAddress := storage.Address{
				Hash:        *trace.Result.Address,
				FirstHeight: decodeCtx.Block.Height,
				LastHeight:  decodeCtx.Block.Height,
				Balance:     storage.EmptyBalance(),
				IsContract:  true,
			}

			var txHash types.Hex
			if trace.TxHash != nil {
				txHash = *trace.TxHash
			}

			contract := &storage.Contract{
				Height:  decodeCtx.Block.Height,
				Address: contractAddress,
				Code:    *trace.Result.Code,
				Tx: &storage.Tx{
					Hash: txHash,
				},
				Deployer: &deployerAddress,
			}

			newTrace.Contract = contract
			decodeCtx.AddAddress(&deployerAddress)
			decodeCtx.AddAddress(&contractAddress)
			decodeCtx.AddContract(contract)
			if parseErr := p.parseProxyContract(decodeCtx, contract); parseErr != nil {
				return parseErr
			}

			if parseErr := ParseEvmContractMetadata(contract); parseErr != nil {
				p.Log.Err(parseErr).
					Str("contract", contract.Address.Hash.String()).
					Uint64("height", uint64(contract.Height)).
					Msg("parsing contract metadata")
			}
		}

		decodeCtx.Block.Traces[i] = newTrace
		decodeCtx.AddTrace(newTrace)
	}

	if err = p.parseTxs(decodeCtx); err != nil {
		return err
	}

	if err = p.parseTransfers(decodeCtx); err != nil {
		return err
	}

	output := p.MustOutput(OutputName)
	output.Push(decodeCtx)

	return nil
}

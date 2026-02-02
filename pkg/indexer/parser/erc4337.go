package parser

import (
	"bytes"
	"math/big"

	"github.com/baking-bad/noble-indexer/internal/storage"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/erc4337"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (p *Module) parseERC4337(ctx *dCtx.Context, tx *storage.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	version, ok := isEntryPointCall(tx)
	if !ok {
		return nil
	}

	switch version {
	case "v0.6":
		decodedParams, err := decodeHandleOps[erc4337.UserOperation](tx, p.abi[erc4337.ABIEntryPointV06])
		if err != nil {
			return errors.Wrap(err, "decoding handleOps params")
		}
		decodedUserOpEvents, err := decodeUserOperationEvents(p.abi[erc4337.ABIEntryPointV06], tx.Logs)
		if err != nil {
			return errors.Wrap(err, "decoding user operation events")
		}
		err = mergeDecodedParamsWithEvents(ctx, tx, decodedParams, decodedUserOpEvents)
		if err != nil {
			return errors.Wrap(err, "merging decoded params with events")
		}

		return nil

	case "v0.7":
		decodedParams, err := decodeHandleOps[erc4337.PackedUserOperation](tx, p.abi[erc4337.ABIEntryPointV07])
		if err != nil {
			return errors.Wrap(err, "decoding v0.7 handleOps params")
		}
		decodedUserOpEvents, err := decodeUserOperationEvents(p.abi[erc4337.ABIEntryPointV07], tx.Logs)
		if err != nil {
			return errors.Wrap(err, "decoding user operation events")
		}
		err = mergeDecodedParamsWithEvents(ctx, tx, decodedParams, decodedUserOpEvents)
		if err != nil {
			return errors.Wrap(err, "merging decoded params with events")
		}

		return nil

	default:
		return errors.Errorf("unsupported EIP-4337 version: %s", version)
	}
}

func isEntryPointCall(tx *storage.Tx) (version string, ok bool) {
	if tx.ToAddress == nil {
		return "", false
	}
	version, ok = erc4337.EntryPointAddresses[tx.ToAddress.Hash.Hex()]
	if !ok {
		return "", false
	}
	if len(tx.Input) < 4 {
		return "", false
	}

	return version, isHandleOpsSelector(tx.Input[:4])
}

func isHandleOpsSelector(selector []byte) bool {
	return bytes.Equal(selector, erc4337.HandleOpsV06Selector) || bytes.Equal(selector, erc4337.HandleOpsV07Selector)
}

func decodeHandleOps[T erc4337.IdentifiableUserOp](
	tx *storage.Tx,
	contractABI *abi.ABI,
) (map[string]T, error) {
	method, err := contractABI.MethodById(tx.Input[:4])
	if err != nil {
		return nil, errors.Wrap(err, "getting method by id")
	}

	args, err := method.Inputs.Unpack(tx.Input[4:])
	if err != nil {
		return nil, errors.Wrap(err, "unpacking inputs")
	}

	if len(args) < 2 {
		return nil, errors.New("invalid args length")
	}

	var ops []T
	converted := abi.ConvertType(args[0], &ops)
	if converted == nil {
		return nil, errors.New("failed to convert user operations")
	}
	operationsMap := make(map[string]T, len(ops))
	for i := range ops {
		operationsMap[ops[i].GetUniqueKey()] = ops[i]
	}

	return operationsMap, nil
}

func decodeUserOperationEvents(
	contractAbi *abi.ABI,
	logs []*storage.Log,
) (map[string]erc4337.UserOperationEvent, error) {
	eventsMap := make(map[string]erc4337.UserOperationEvent)
	for i := range logs {
		if len(logs[i].Topics) < 4 {
			continue
		}
		if logs[i].Topics[0].Hex() != erc4337.UserOperationEventSignature {
			continue
		}

		event := erc4337.UserOperationEvent{}
		event.UserOpHash = common.BytesToHash(logs[i].Topics[1])
		event.Sender = common.BytesToAddress(logs[i].Topics[2])
		event.Paymaster = common.BytesToAddress(logs[i].Topics[3])

		if err := contractAbi.UnpackIntoInterface(&event, "UserOperationEvent", logs[i].Data); err != nil {
			return nil, errors.Wrap(err, "failed to unpack log data")
		}
		eventsMap[event.GetUniqueKey()] = event
	}

	return eventsMap, nil
}

func mergeDecodedParamsWithEvents[T erc4337.CommonUserOperation](
	ctx *dCtx.Context,
	tx *storage.Tx,
	decodedHandleOpsParams map[string]T,
	events map[string]erc4337.UserOperationEvent,
) error {
	if len(decodedHandleOpsParams) == 0 {
		return nil
	}
	if len(decodedHandleOpsParams) != len(events) {
		return errors.New("decoded params length mismatch")
	}

	for key, decodedParams := range decodedHandleOpsParams {
		event, ok := events[key]
		if !ok {
			return errors.New("user operation event not found")
		}

		var accountGasLimits []byte
		var gasFees []byte
		switch op := any(decodedParams).(type) {
		case erc4337.UserOperation:
			accountGasLimits = packUint128Pair(op.VerificationGasLimit, op.CallGasLimit)
			gasFees = packUint128Pair(op.MaxPriorityFeePerGas, op.MaxFeePerGas)

		case erc4337.PackedUserOperation:
			accountGasLimits = op.AccountGasLimits[:]
			gasFees = op.GasFees[:]
		}
		senderAddress := &storage.Address{
			Hash:       decodedParams.GetSender().Bytes(),
			LastHeight: ctx.Block.Height,
			IsContract: true,
			Balance:    storage.EmptyBalance(),
		}
		senderContract := &storage.Contract{
			Address: *senderAddress,
		}
		ctx.AddContract(senderContract)
		ctx.AddAddress(senderAddress)

		var paymasterAddress *storage.Address
		if event.Paymaster != (common.Address{}) {
			paymasterAddress = &storage.Address{
				Hash:       event.Paymaster.Bytes(),
				LastHeight: ctx.Block.Height,
				IsContract: true,
				Balance:    storage.EmptyBalance(),
			}
			paymasterContract := &storage.Contract{
				Address: *paymasterAddress,
			}
			ctx.AddContract(paymasterContract)
			ctx.AddAddress(paymasterAddress)
		}

		ctx.AddAddress(&tx.FromAddress)
		ctx.AddUserOp(&storage.ERC4337UserOp{
			Time:               tx.Time,
			Height:             tx.Height,
			Hash:               event.UserOpHash[:],
			Nonce:              decimal.NewFromBigInt(decodedParams.GetNonce(), 0),
			Success:            event.Success,
			ActualGasCost:      decimal.NewFromBigInt(event.ActualGasCost, 0),
			ActualGasUsed:      decimal.NewFromBigInt(event.ActualGasUsed, 0),
			InitCode:           decodedParams.GetInitCode(),
			CallData:           decodedParams.GetCallData(),
			AccountGasLimits:   accountGasLimits,
			PreVerificationGas: decimal.NewFromBigInt(decodedParams.GetPreVerificationGas(), 0),
			GasFees:            gasFees,
			PaymasterAndData:   decodedParams.GetPaymasterAndData(),
			Signature:          decodedParams.GetSignature(),
			Tx:                 *tx,
			Sender:             *senderAddress,
			Paymaster:          paymasterAddress,
			Bundler:            tx.FromAddress,
		})
	}

	return nil
}

func packUint128Pair(high, low *big.Int) []byte {
	result := make([]byte, 32)
	if high != nil {
		b := high.Bytes()
		copy(result[16-len(b):16], b)
	}
	if low != nil {
		b := low.Bytes()
		copy(result[32-len(b):32], b)
	}
	return result
}

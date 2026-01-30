package eip4337

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (op PackedUserOperation) GetSender() common.Address {
	return op.Sender
}

func (op PackedUserOperation) GetNonce() *big.Int {
	return op.Nonce
}

func (op PackedUserOperation) GetInitCode() []byte {
	return op.InitCode
}

func (op PackedUserOperation) GetCallData() []byte {
	return op.CallData
}

func (op PackedUserOperation) GetPreVerificationGas() *big.Int {
	return op.PreVerificationGas
}

func (op PackedUserOperation) GetPaymasterAndData() []byte {
	return op.PaymasterAndData
}

func (op PackedUserOperation) GetSignature() []byte {
	return op.Signature
}

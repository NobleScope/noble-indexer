package eip4337

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (op UserOperation) GetSender() common.Address {
	return op.Sender
}

func (op UserOperation) GetNonce() *big.Int {
	return op.Nonce
}

func (op UserOperation) GetInitCode() []byte {
	return op.InitCode
}

func (op UserOperation) GetCallData() []byte {
	return op.CallData
}

func (op UserOperation) GetPreVerificationGas() *big.Int {
	return op.PreVerificationGas
}

func (op UserOperation) GetPaymasterAndData() []byte {
	return op.PaymasterAndData
}

func (op UserOperation) GetSignature() []byte {
	return op.Signature
}

package responses

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/shopspring/decimal"
)

// UserOp model info
//
//	@Description	ERC-4337 user operation information
type UserOp struct {
	Id                 uint64          `example:"1"                                                                  json:"id"                   swaggertype:"integer"`
	Height             uint64          `example:"100"                                                                json:"height"               swaggertype:"integer"`
	Time               time.Time       `example:"2026-01-01T01:01:01+00:00"                                          format:"date-time"          json:"time"           swaggertype:"string"`
	TxHash             string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"tx_hash"              swaggertype:"string"`
	Hash               string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"hash"                 swaggertype:"string"`
	Sender             string          `example:"0x0000000000000000000000000000000000000001"                         json:"sender"               swaggertype:"string"`
	Paymaster          *string         `example:"0x0000000000000000000000000000000000000002"                         json:"paymaster"            swaggertype:"string"`
	Bundler            string          `example:"0x0000000000000000000000000000000000000003"                         json:"bundler"              swaggertype:"string"`
	Nonce              decimal.Decimal `example:"1"                                                                  json:"nonce"                swaggertype:"string"`
	Success            bool            `example:"true"                                                               json:"success"              swaggertype:"boolean"`
	ActualGasCost      decimal.Decimal `example:"21000"                                                              json:"actual_gas_cost"      swaggertype:"string"`
	ActualGasUsed      decimal.Decimal `example:"21000"                                                              json:"actual_gas_used"      swaggertype:"string"`
	InitCode           string          `example:"0x"                                                                 json:"init_code"            swaggertype:"string"`
	CallData           string          `example:"0x"                                                                 json:"call_data"            swaggertype:"string"`
	AccountGasLimits   string          `example:"0x"                                                                 json:"account_gas_limits"   swaggertype:"string"`
	PreVerificationGas decimal.Decimal `example:"21000"                                                              json:"pre_verification_gas" swaggertype:"string"`
	GasFees            string          `example:"0x"                                                                 json:"gas_fees"             swaggertype:"string"`
	PaymasterAndData   string          `example:"0x"                                                                 json:"paymaster_and_data"   swaggertype:"string"`
	Signature          string          `example:"0x"                                                                 json:"signature"            swaggertype:"string"`
}

func NewUserOp(op storage.ERC4337UserOp) UserOp {
	result := UserOp{
		Id:                 op.Id,
		Height:             uint64(op.Height),
		Time:               op.Time,
		TxHash:             op.Tx.Hash.Hex(),
		Hash:               op.Hash.Hex(),
		Sender:             op.Sender.Hash.Hex(),
		Bundler:            op.Bundler.Hash.Hex(),
		Nonce:              op.Nonce,
		Success:            op.Success,
		ActualGasCost:      op.ActualGasCost,
		ActualGasUsed:      op.ActualGasUsed,
		InitCode:           op.InitCode.Hex(),
		CallData:           op.CallData.Hex(),
		AccountGasLimits:   op.AccountGasLimits.Hex(),
		PreVerificationGas: op.PreVerificationGas,
		GasFees:            op.GasFees.Hex(),
		PaymasterAndData:   op.PaymasterAndData.Hex(),
		Signature:          op.Signature.Hex(),
	}

	if op.Paymaster != nil {
		paymaster := op.Paymaster.Hash.Hex()
		result.Paymaster = &paymaster
	}

	return result
}

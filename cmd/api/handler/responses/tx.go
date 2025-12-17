package responses

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	Height            uint64          `example:"100"                                                                json:"height"              swaggertype:"integer"`
	Time              time.Time       `example:"2023-07-04T03:10:57+00:00"                                          json:"time"                swaggertype:"string"`
	Hash              string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"hash"                swaggertype:"string"`
	Index             int64           `example:"0"                                                                  json:"index"               swaggertype:"integer"`
	Nonce             int64           `example:"1"                                                                  json:"nonce"               swaggertype:"integer"`
	Type              string          `enums:"TxTypeUnknown,TxTypeLegacy,TxTypeDynamicFee,TxTypeBlob,TxTypeSetCode" example:"TxTypeDynamicFee" json:"type"           swaggertype:"string"`
	Status            string          `enums:"TxStatusSuccess,TxStatusRevert"                                       example:"TxStatusSuccess"  json:"status"         swaggertype:"string"`
	Gas               decimal.Decimal `example:"21000"                                                              json:"gas"                 swaggertype:"integer"`
	GasPrice          decimal.Decimal `example:"1000000"                                                            json:"gas_price"           swaggertype:"integer"`
	GasUsed           decimal.Decimal `example:"21000"                                                              json:"gas_used"            swaggertype:"integer"`
	CumulativeGasUsed decimal.Decimal `example:"21000"                                                              json:"cumulative_gas_used" swaggertype:"integer"`
	EffectiveGasPrice decimal.Decimal `example:"1000000"                                                            json:"effective_gas_price" swaggertype:"integer"`
	Fee               decimal.Decimal `example:"21000000000"                                                        json:"fee"                 swaggertype:"string"`
	Amount            decimal.Decimal `example:"1000000000000000000"                                                json:"amount"              swaggertype:"integer"`
	FromAddress       string          `example:"0x0000000000000000000000000000000000000000"                         json:"from_address"        swaggertype:"string"`
	ToAddress         *string         `example:"0x0000000000000000000000000000000000000001"                         json:"to_address"          swaggertype:"string"`
	Input             string          `example:"0x"                                                                 json:"input"               swaggertype:"string"`
	LogsBloom         string          `example:"0x00000000000000000000000000000000000000000000"                     json:"logs_bloom"          swaggertype:"string"`
}

func NewTransaction(tx storage.Tx) Transaction {
	result := Transaction{
		Height:            uint64(tx.Height),
		Time:              tx.Time,
		Hash:              tx.Hash.Hex(),
		Index:             tx.Index,
		Nonce:             tx.Nonce,
		Type:              string(tx.Type),
		Status:            string(tx.Status),
		Gas:               tx.Gas,
		GasPrice:          tx.GasPrice,
		GasUsed:           tx.GasUsed,
		CumulativeGasUsed: tx.CumulativeGasUsed,
		EffectiveGasPrice: tx.EffectiveGasPrice,
		Fee:               tx.Fee,
		Amount:            tx.Amount,
		FromAddress:       tx.FromAddress.String(),
		Input:             types.Hex(tx.Input).Hex(),
		LogsBloom:         types.Hex(tx.LogsBloom).Hex(),
	}

	if tx.ToAddress != nil {
		toAddr := tx.ToAddress.String()
		result.ToAddress = &toAddr
	}

	return result
}

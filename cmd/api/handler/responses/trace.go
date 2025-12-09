package responses

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/shopspring/decimal"
)

type Trace struct {
	Height         uint64           `example:"100"                                                    json:"height"          swaggertype:"integer"`
	Time           time.Time        `example:"2023-07-04T03:10:57+00:00"                              json:"time"            swaggertype:"string"`
	TxHash         string           `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22"       json:"tx_hash"         swaggertype:"string"`
	FromAddress    string           `example:"0x0000000000000000000000000000000000000000"             json:"from_address"    swaggertype:"string"`
	ToAddress      *string          `example:"0x0000000000000000000000000000000000000001"             json:"to_address"      swaggertype:"string"`
	GasLimit       decimal.Decimal  `example:"21000"                                                  json:"gas_limit"       swaggertype:"string"`
	Amount         *decimal.Decimal `example:"1000000000000000000"                                    json:"amount"          swaggertype:"string"`
	Input          *string          `example:"0x0"                                                    json:"input"           swaggertype:"string"`
	TxPosition     uint64           `example:"0"                                                      json:"tx_position"     swaggertype:"integer"`
	TraceAddress   []uint64         `example:"0,1"                                                    json:"trace_address"   swaggertype:"array,integer"`
	Type           string           `enums:"call,delegatecall,staticcall,create,create2,selfdestruct" example:"call"         json:"type"                 swaggertype:"string"`
	InitHash       *string          `example:"0x6060604052341561000f57600080fd5b"                     json:"init_hash"       swaggertype:"string"`
	CreationMethod *string          `example:"create"                                                 json:"creation_method" swaggertype:"string"`
	GasUsed        decimal.Decimal  `example:"21000"                                                  json:"gas_used"        swaggertype:"string"`
	Output         *string          `example:"0x0"                                                    json:"output"          swaggertype:"string"`
	Contract       *string          `example:"0x0000000000000000000000000000000000000002"             json:"contract"        swaggertype:"string"`
	Subtraces      uint64           `example:"0"                                                      json:"subtraces"       swaggertype:"integer"`
}

func NewTrace(t *storage.Trace) Trace {
	result := Trace{
		Height:         uint64(t.Height),
		Time:           t.Time,
		TxHash:         t.Tx.Hash.Hex(),
		FromAddress:    t.FromAddress.String(),
		GasLimit:       t.GasLimit,
		Amount:         t.Amount,
		TxPosition:     t.TxPosition,
		TraceAddress:   t.TraceAddress,
		Type:           string(t.Type),
		CreationMethod: t.CreationMethod,
		GasUsed:        t.GasUsed,
		Subtraces:      t.Subtraces,
	}

	if t.ToAddress != nil {
		toAddr := t.ToAddress.String()
		result.ToAddress = &toAddr
	}
	if t.InitHash != nil {
		initHash := t.InitHash.String()
		result.InitHash = &initHash
	}
	if t.Input != nil {
		input := types.Hex(t.Input).String()
		result.Input = &input
	}
	if t.Output != nil {
		output := types.Hex(t.Output).String()
		result.Output = &output
	}
	if t.Contract != nil {
		contract := t.Contract.String()
		result.Contract = &contract
	}

	return result
}

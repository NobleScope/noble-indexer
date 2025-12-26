package responses

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/shopspring/decimal"
)

// Trace model info
//
//	@Description	Transaction execution trace information
type Trace struct {
	Height         uint64           `example:"100"                                                           json:"height"          swaggertype:"integer"`
	Time           time.Time        `example:"2023-07-04T03:10:57+00:00"                                     json:"time"            swaggertype:"string"`
	TxHash         *string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22"              json:"tx_hash"         swaggertype:"string"`
	FromAddress    *string          `example:"0x0000000000000000000000000000000000000000"                    json:"from_address"    swaggertype:"string"`
	ToAddress      *string          `example:"0x0000000000000000000000000000000000000001"                    json:"to_address"      swaggertype:"string"`
	GasLimit       decimal.Decimal  `example:"21000"                                                         json:"gas_limit"       swaggertype:"string"`
	Amount         *decimal.Decimal `example:"1000000000000000000"                                           json:"amount"          swaggertype:"string"`
	Input          *string          `example:"0x0"                                                           json:"input"           swaggertype:"string"`
	TxPosition     uint64           `example:"0"                                                             json:"tx_position"     swaggertype:"integer"`
	TraceAddress   []uint64         `example:"0,1"                                                           json:"trace_address"   swaggertype:"array,integer"`
	Type           string           `enums:"call,delegatecall,staticcall,create,create2,selfdestruct,reward" example:"call"         json:"type"                 swaggertype:"string"`
	InitHash       *string          `example:"0x6060604052341561000f57600080fd5b"                            json:"init_hash"       swaggertype:"string"`
	CreationMethod *string          `example:"create"                                                        json:"creation_method" swaggertype:"string"`
	GasUsed        decimal.Decimal  `example:"21000"                                                         json:"gas_used"        swaggertype:"string"`
	Output         *string          `example:"0x0"                                                           json:"output"          swaggertype:"string"`
	Contract       *string          `example:"0x0000000000000000000000000000000000000002"                    json:"contract"        swaggertype:"string"`
	Subtraces      uint64           `example:"0"                                                             json:"subtraces"       swaggertype:"integer"`
}

func NewTrace(t *storage.Trace) Trace {
	result := Trace{
		Height:         uint64(t.Height),
		Time:           t.Time,
		GasLimit:       t.GasLimit,
		Amount:         t.Amount,
		TraceAddress:   t.TraceAddress,
		Type:           string(t.Type),
		CreationMethod: t.CreationMethod,
		GasUsed:        t.GasUsed,
		Subtraces:      t.Subtraces,
	}

	if t.Tx != nil && t.Tx.Hash != nil {
		txHash := t.Tx.Hash.Hex()
		result.TxHash = &txHash
	}
	if t.ToAddress != nil {
		toAddr := t.ToAddress.Hash.Hex()
		result.ToAddress = &toAddr
	}
	if t.FromAddress != nil {
		fromAddr := t.FromAddress.Hash.Hex()
		result.FromAddress = &fromAddr
	}
	if t.InitHash != nil {
		initHash := t.InitHash.Hex()
		result.InitHash = &initHash
	}
	if t.Input != nil {
		input := types.Hex(t.Input).Hex()
		result.Input = &input
	}
	if t.Output != nil {
		output := types.Hex(t.Output).Hex()
		result.Output = &output
	}
	if t.Contract != nil {
		contract := t.Contract.Address.Hash.Hex()
		result.Contract = &contract
	}
	if t.TxPosition != nil {
		result.TxPosition = *t.TxPosition
	}

	return result
}

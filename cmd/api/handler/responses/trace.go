package responses

import (
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// Trace model info
//
//	@Description	Transaction execution trace information
type Trace struct {
	Height         uint64           `example:"100"                                                           json:"height"                    swaggertype:"integer"`
	Time           time.Time        `example:"2023-07-04T03:10:57+00:00"                                     json:"time"                      swaggertype:"string"`
	TxHash         *string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22"              json:"tx_hash,omitempty"         swaggertype:"string"`
	FromAddress    *string          `example:"0x0000000000000000000000000000000000000001"                    json:"from_address,omitempty"    swaggertype:"string"`
	ToAddress      *string          `example:"0x123456789abcdef123456789abcdef123456789abc"                  json:"to_address,omitempty"      swaggertype:"string"`
	GasLimit       decimal.Decimal  `example:"2100"                                                          json:"gas_limit"                 swaggertype:"string"`
	Amount         *decimal.Decimal `example:"123456789123456789"                                            json:"amount,omitempty"          swaggertype:"string"`
	Input          *string          `example:"hex input data"                                                json:"input,omitempty"           swaggertype:"string"`
	TxPosition     uint64           `example:"123456789"                                                     json:"tx_position"               swaggertype:"integer"`
	TraceAddress   []uint64         `example:"1,2,3"                                                         json:"trace_address"             swaggertype:"array,integer"`
	Type           string           `enums:"call,delegatecall,staticcall,create,create2,selfdestruct,reward" example:"call"                   json:"type"                 swaggertype:"string"`
	InitHash       *string          `example:"0x6060604052341561000f57600080fd5b"                            json:"init_hash,omitempty"       swaggertype:"string"`
	CreationMethod *string          `example:"create"                                                        json:"creation_method,omitempty" swaggertype:"string"`
	GasUsed        decimal.Decimal  `example:"21000"                                                         json:"gas_used"                  swaggertype:"string"`
	Output         *string          `example:"0x0"                                                           json:"output,omitempty"          swaggertype:"string"`
	Contract       *string          `example:"0x0000000000000000000000000000000000000002"                    json:"contract,omitempty"        swaggertype:"string"`
	Subtraces      uint64           `example:"0"                                                             json:"subtraces"                 swaggertype:"integer"`
	Decoded        *DecodedTrace    `json:"decoded,omitempty"                                                swaggertype:"object"`
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
	if t.To != nil {
		if parsedABI := parseABI(t.ToContractABI); parsedABI != nil {
			result.Decoded = decodeTxArgs(parsedABI, t.Input)
		}
	}

	return result
}

var errInvalidTraceAddress = errors.New("invalid trace address")

// TraceTreeItem model info
//
//	@Description	Transaction execution trace tree item
type TraceTreeItem struct {
	*Trace
	Children []*TraceTreeItem `json:"children,omitempty" swaggertype:"array,object"`
}

func BuildTraceTree(traces []*storage.Trace) (*TraceTreeItem, error) {
	var root *TraceTreeItem

	appendChildren := func(parent *TraceTreeItem, child *TraceTreeItem, idx uint64) error {
		if parent == nil {
			return errors.Wrap(errInvalidTraceAddress, "root is nil")
		}
		if idx >= uint64(len(parent.Children)) {
			return errors.Wrap(errInvalidTraceAddress, "trace address out of range")
		}
		parent.Children[idx] = child
		return nil
	}

	for i := range traces {
		if traces[i] == nil {
			return nil, errors.New("nil trace in traces list")
		}
		resp := NewTrace(traces[i])

		item := TraceTreeItem{
			Trace:    &resp,
			Children: make([]*TraceTreeItem, resp.Subtraces),
		}

		switch len(resp.TraceAddress) {
		case 0:
			// root trace
			root = &item
		case 1:
			// first level trace
			if err := appendChildren(root, &item, resp.TraceAddress[0]); err != nil {
				return nil, err
			}
		default:
			// deeper levels
			current := root
			for j := 0; j < len(resp.TraceAddress)-1; j++ {
				if current == nil {
					return nil, errors.Wrap(errInvalidTraceAddress, "current is nil")
				}
				if resp.TraceAddress[j] >= uint64(len(current.Children)) {
					return nil, errors.Wrap(errInvalidTraceAddress, "trace address out of range")
				}
				current = current.Children[resp.TraceAddress[j]]
			}

			if err := appendChildren(current, &item, resp.TraceAddress[len(resp.TraceAddress)-1]); err != nil {
				return nil, err
			}
		}
	}
	return root, nil
}

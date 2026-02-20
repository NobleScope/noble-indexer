package trace_provider

import (
	stdjson "encoding/json"
	"strings"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type GethTraceResult struct {
	TxHash pkgTypes.Hex     `json:"txHash"`
	Result GethCallResponse `json:"result"`
}

type GethCallResponse struct {
	Type    string             `json:"type"`
	From    pkgTypes.Hex       `json:"from"`
	To      *pkgTypes.Hex      `json:"to"`
	Value   *pkgTypes.Hex      `json:"value"`
	Gas     pkgTypes.Hex       `json:"gas"`
	GasUsed pkgTypes.Hex       `json:"gasUsed"`
	Input   pkgTypes.Hex       `json:"input"`
	Output  *pkgTypes.Hex      `json:"output"`
	Error   *string            `json:"error"`
	Calls   []GethCallResponse `json:"calls"`
	Address *pkgTypes.Hex      `json:"address"` // created contract address (CREATE/CREATE2)
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	gethTraceTypeCall         = "call"
	gethTraceTypeCreate       = "create"
	gethTraceTypeCreate2      = "create2"
	gethTraceTypeSelfdestruct = "selfdestruct"
	gethTraceTypeSuicide      = "suicide"
	gethCallTypeCall          = "call"
	gethCallTypeStaticcall    = "staticcall"
	gethCallTypeDelegatecall  = "delegatecall"
	gethCallTypeCallcode      = "callcode"
)

func mapGethType(gethType string) (string, *string) {
	upper := strings.ToUpper(gethType)
	switch upper {
	case "CALL":
		ct := gethCallTypeCall
		return gethTraceTypeCall, &ct
	case "STATICCALL":
		ct := gethCallTypeStaticcall
		return gethTraceTypeCall, &ct
	case "DELEGATECALL":
		ct := gethCallTypeDelegatecall
		return gethTraceTypeCall, &ct
	case "CALLCODE":
		ct := gethCallTypeCallcode
		return gethTraceTypeCall, &ct
	case "CREATE":
		return gethTraceTypeCreate, nil
	case "CREATE2":
		return gethTraceTypeCreate2, nil
	case "SELFDESTRUCT":
		return gethTraceTypeSelfdestruct, nil
	case "SUICIDE":
		return gethTraceTypeSuicide, nil
	default:
		return strings.ToLower(gethType), nil
	}
}

type GethDebugTraceProvider struct{}

func (g *GethDebugTraceProvider) Method() string {
	return "debug_traceBlockByNumber"
}

func (g *GethDebugTraceProvider) Params(hexLevel string) []any {
	return []any{
		hexLevel,
		map[string]string{"tracer": "callTracer"},
	}
}

func (g *GethDebugTraceProvider) ParseTraces(raw stdjson.RawMessage) ([]pkgTypes.Trace, error) {
	var results []GethTraceResult
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, errors.Wrap(err, "unmarshal geth trace results")
	}

	var traces []pkgTypes.Trace
	for txPos, result := range results {
		txPosition := uint64(txPos)
		txHash := result.TxHash
		flatTraces := flattenGethCallFrame(result.Result, &txHash, &txPosition, []uint64{})
		traces = append(traces, flatTraces...)
	}

	return traces, nil
}

func flattenGethCallFrame(frame GethCallResponse, txHash *pkgTypes.Hex, txPosition *uint64, traceAddress []uint64) []pkgTypes.Trace {
	traceType, callType := mapGethType(frame.Type)

	from := frame.From
	trace := pkgTypes.Trace{
		Action: pkgTypes.Action{
			From:     &from,
			To:       frame.To,
			Gas:      &frame.Gas,
			Value:    frame.Value,
			Input:    &frame.Input,
			CallType: callType,
		},
		Result: pkgTypes.TraceResult{
			GasUsed: frame.GasUsed,
			Output:  frame.Output,
		},
		TxHash:       txHash,
		TxPosition:   txPosition,
		TraceAddress: traceAddress,
		Subtraces:    uint64(len(frame.Calls)),
		Type:         traceType,
		Error:        frame.Error,
	}

	if traceType == gethTraceTypeCreate || traceType == gethTraceTypeCreate2 {
		trace.Action.Input = nil
		trace.Action.Init = &frame.Input
		trace.Action.To = nil
		trace.Action.CreationMethod = &traceType
		trace.Result.Address = frame.Address
		trace.Result.Output = nil
		if frame.Output != nil {
			trace.Result.Code = frame.Output
		}
	}

	result := []pkgTypes.Trace{trace}

	for i, child := range frame.Calls {
		childAddress := make([]uint64, len(traceAddress)+1)
		copy(childAddress, traceAddress)
		childAddress[len(traceAddress)] = uint64(i)
		result = append(result, flattenGethCallFrame(child, txHash, txPosition, childAddress)...)
	}

	return result
}

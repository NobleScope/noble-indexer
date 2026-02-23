package trace_provider

import (
	stdjson "encoding/json"
	"testing"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestMapGethType(t *testing.T) {
	tests := []struct {
		input        string
		wantType     string
		wantCallType *string
	}{
		{"CALL", "call", strPtr("call")},
		{"STATICCALL", "call", strPtr("staticcall")},
		{"DELEGATECALL", "call", strPtr("delegatecall")},
		{"CALLCODE", "call", strPtr("callcode")},
		{"CREATE", "create", nil},
		{"CREATE2", "create2", nil},
		{"SELFDESTRUCT", "selfdestruct", nil},
		{"SUICIDE", "suicide", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotType, gotCallType := mapGethType(tt.input)
			require.Equal(t, tt.wantType, gotType)
			if tt.wantCallType == nil {
				require.Nil(t, gotCallType)
			} else {
				require.NotNil(t, gotCallType)
				require.Equal(t, *tt.wantCallType, *gotCallType)
			}
		})
	}
}

func TestFlattenGethCallFrame_SimpleCall(t *testing.T) {
	txHash := mustHex("0xaabbccdd")
	txPos := uint64(0)

	frame := GethCallResponse{
		Type:    "CALL",
		From:    mustHex("0x1111111111111111111111111111111111111111"),
		To:      hexPtr("0x2222222222222222222222222222222222222222"),
		Value:   hexPtr("0x100"),
		Gas:     mustHex("0x5208"),
		GasUsed: mustHex("0x5208"),
		Input:   mustHex("0x"),
		Output:  hexPtr("0x"),
	}

	traces := flattenGethCallFrame(frame, &txHash, &txPos, []uint64{})

	require.Len(t, traces, 1)
	tr := traces[0]

	require.Equal(t, "call", tr.Type)
	require.NotNil(t, tr.Action.CallType)
	require.Equal(t, "call", *tr.Action.CallType)
	require.NotNil(t, tr.Action.From)
	require.NotNil(t, tr.Action.To)
	require.NotNil(t, tr.Action.Gas)
	require.NotNil(t, tr.Action.Value)
	require.Equal(t, uint64(0), tr.Subtraces)
	require.Empty(t, tr.TraceAddress)
	require.Equal(t, &txHash, tr.TxHash)
	require.Equal(t, &txPos, tr.TxPosition)
}

func TestFlattenGethCallFrame_NestedCalls(t *testing.T) {
	txHash := mustHex("0xaabb")
	txPos := uint64(1)

	frame := GethCallResponse{
		Type:    "CALL",
		From:    mustHex("0x1111111111111111111111111111111111111111"),
		To:      hexPtr("0x2222222222222222222222222222222222222222"),
		Gas:     mustHex("0xffff"),
		GasUsed: mustHex("0xaaaa"),
		Input:   mustHex("0x"),
		Calls: []GethCallResponse{
			{
				Type:    "STATICCALL",
				From:    mustHex("0x2222222222222222222222222222222222222222"),
				To:      hexPtr("0x3333333333333333333333333333333333333333"),
				Gas:     mustHex("0x1000"),
				GasUsed: mustHex("0x500"),
				Input:   mustHex("0xabcd"),
				Output:  hexPtr("0x1234"),
			},
			{
				Type:    "DELEGATECALL",
				From:    mustHex("0x2222222222222222222222222222222222222222"),
				To:      hexPtr("0x4444444444444444444444444444444444444444"),
				Gas:     mustHex("0x2000"),
				GasUsed: mustHex("0x1000"),
				Input:   mustHex("0x"),
				Calls: []GethCallResponse{
					{
						Type:    "CALL",
						From:    mustHex("0x4444444444444444444444444444444444444444"),
						To:      hexPtr("0x5555555555555555555555555555555555555555"),
						Gas:     mustHex("0x800"),
						GasUsed: mustHex("0x400"),
						Input:   mustHex("0x"),
					},
				},
			},
		},
	}

	traces := flattenGethCallFrame(frame, &txHash, &txPos, []uint64{})

	require.Len(t, traces, 4)

	// Root: traceAddress=[], subtraces=2
	require.Empty(t, traces[0].TraceAddress)
	require.Equal(t, uint64(2), traces[0].Subtraces)
	require.Equal(t, "call", traces[0].Type)
	require.Equal(t, "call", *traces[0].Action.CallType)

	// First child: traceAddress=[0], subtraces=0
	require.Equal(t, []uint64{0}, traces[1].TraceAddress)
	require.Equal(t, uint64(0), traces[1].Subtraces)
	require.Equal(t, "call", traces[1].Type)
	require.Equal(t, "staticcall", *traces[1].Action.CallType)

	// Second child: traceAddress=[1], subtraces=1
	require.Equal(t, []uint64{1}, traces[2].TraceAddress)
	require.Equal(t, uint64(1), traces[2].Subtraces)
	require.Equal(t, "call", traces[2].Type)
	require.Equal(t, "delegatecall", *traces[2].Action.CallType)

	// Grandchild: traceAddress=[1,0], subtraces=0
	require.Equal(t, []uint64{1, 0}, traces[3].TraceAddress)
	require.Equal(t, uint64(0), traces[3].Subtraces)
	require.Equal(t, "call", traces[3].Type)
	require.Equal(t, "call", *traces[3].Action.CallType)
}

func TestFlattenGethCallFrame_CreateTrace(t *testing.T) {
	txHash := mustHex("0xdead")
	txPos := uint64(0)

	contractAddr := mustHex("0x9999999999999999999999999999999999999999")
	bytecode := mustHex("0x6080604052")

	frame := GethCallResponse{
		Type:    "CREATE",
		From:    mustHex("0x1111111111111111111111111111111111111111"),
		Value:   hexPtr("0x0"),
		Gas:     mustHex("0x10000"),
		GasUsed: mustHex("0x8000"),
		Input:   mustHex("0x6080604052348015"),
		Output:  &bytecode,
		Address: &contractAddr,
	}

	traces := flattenGethCallFrame(frame, &txHash, &txPos, []uint64{})

	require.Len(t, traces, 1)
	tr := traces[0]

	require.Equal(t, "create", tr.Type)
	require.Nil(t, tr.Action.CallType)
	require.Nil(t, tr.Action.To)
	require.NotNil(t, tr.Action.Init)
	require.NotNil(t, tr.Action.CreationMethod)
	require.Equal(t, "create", *tr.Action.CreationMethod)
	require.Nil(t, tr.Result.Output)
	require.NotNil(t, tr.Result.Address)
	require.NotNil(t, tr.Result.Code)
	require.Equal(t, contractAddr, *tr.Result.Address)
	require.Equal(t, bytecode, *tr.Result.Code)
}

func TestFlattenGethCallFrame_SelfdestructTrace(t *testing.T) {
	txHash := mustHex("0xdead")
	txPos := uint64(0)

	frame := GethCallResponse{
		Type:  "SELFDESTRUCT",
		From:  mustHex("0x1111111111111111111111111111111111111111"),
		To:    hexPtr("0x2222222222222222222222222222222222222222"),
		Value: hexPtr("0x1bc16d674ec80000"),
	}

	traces := flattenGethCallFrame(frame, &txHash, &txPos, []uint64{})

	require.Len(t, traces, 1)
	tr := traces[0]

	require.Equal(t, "selfdestruct", tr.Type)
	require.Nil(t, tr.Action.CallType)
	require.NotNil(t, tr.Action.From)
	require.NotNil(t, tr.Action.To)
	require.NotNil(t, tr.Action.Value)
	require.Nil(t, tr.Action.Gas)
	require.Nil(t, tr.Action.Input)
	require.Equal(t, uint64(0), tr.Subtraces)
}

func TestFlattenGethCallFrame_ErrorTrace(t *testing.T) {
	txHash := mustHex("0xbeef")
	txPos := uint64(0)
	errMsg := "execution reverted"

	frame := GethCallResponse{
		Type:    "CALL",
		From:    mustHex("0x1111111111111111111111111111111111111111"),
		To:      hexPtr("0x2222222222222222222222222222222222222222"),
		Gas:     mustHex("0x5208"),
		GasUsed: mustHex("0x0"),
		Input:   mustHex("0x"),
		Error:   &errMsg,
	}

	traces := flattenGethCallFrame(frame, &txHash, &txPos, []uint64{})

	require.Len(t, traces, 1)
	require.NotNil(t, traces[0].Error)
	require.Equal(t, "execution reverted", *traces[0].Error)
}

func TestGethDebugTraceProvider_ParseTraces(t *testing.T) {
	raw := stdjson.RawMessage(`[
		{
			"txHash": "0xaabb",
			"result": {
				"type": "CALL",
				"from": "0x1111111111111111111111111111111111111111",
				"to": "0x2222222222222222222222222222222222222222",
				"value": "0x0",
				"gas": "0x5208",
				"gasUsed": "0x5208",
				"input": "0x",
				"output": "0x",
				"calls": [
					{
						"type": "STATICCALL",
						"from": "0x2222222222222222222222222222222222222222",
						"to": "0x3333333333333333333333333333333333333333",
						"gas": "0x1000",
						"gasUsed": "0x500",
						"input": "0xabcd",
						"output": "0x1234"
					}
				]
			}
		},
		{
			"txHash": "0xccdd",
			"result": {
				"type": "CALL",
				"from": "0x4444444444444444444444444444444444444444",
				"to": "0x5555555555555555555555555555555555555555",
				"value": "0x100",
				"gas": "0x7530",
				"gasUsed": "0x3000",
				"input": "0x",
				"output": "0x"
			}
		}
	]`)

	provider := &GethDebugTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)

	// 2 root traces + 1 child = 3 total
	require.Len(t, traces, 3)

	// First tx root
	require.Equal(t, "call", traces[0].Type)
	require.Equal(t, uint64(1), traces[0].Subtraces)
	require.NotNil(t, traces[0].TxPosition)
	require.Equal(t, uint64(0), *traces[0].TxPosition)

	// First tx child
	require.Equal(t, "call", traces[1].Type)
	require.Equal(t, "staticcall", *traces[1].Action.CallType)
	require.Equal(t, []uint64{0}, traces[1].TraceAddress)
	require.NotNil(t, traces[1].TxPosition)
	require.Equal(t, uint64(0), *traces[1].TxPosition)

	// Second tx root
	require.Equal(t, "call", traces[2].Type)
	require.Equal(t, uint64(0), traces[2].Subtraces)
	require.NotNil(t, traces[2].TxPosition)
	require.Equal(t, uint64(1), *traces[2].TxPosition)
}

func TestGethDebugTraceProvider_MethodAndParams(t *testing.T) {
	provider := &GethDebugTraceProvider{}

	require.Equal(t, "debug_traceBlockByNumber", provider.Method())

	params := provider.Params("0x1a4")
	require.Len(t, params, 2)
	require.Equal(t, "0x1a4", params[0])
	tracerCfg, ok := params[1].(map[string]string)
	require.True(t, ok)
	require.Equal(t, "callTracer", tracerCfg["tracer"])
}

func TestParityTraceProvider_MethodAndParams(t *testing.T) {
	provider := &ParityTraceProvider{}

	require.Equal(t, "trace_block", provider.Method())

	params := provider.Params("0x1a4")
	require.Len(t, params, 1)
	require.Equal(t, "0x1a4", params[0])
}

// helpers

func mustHex(s string) pkgTypes.Hex {
	h, err := pkgTypes.HexFromString(s)
	if err != nil {
		panic(err)
	}
	return h
}

func hexPtr(s string) *pkgTypes.Hex {
	h := mustHex(s)
	return &h
}

func strPtr(s string) *string {
	return &s
}

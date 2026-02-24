package trace_provider

import (
	stdjson "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParityTraceProvider_Method(t *testing.T) {
	provider := &ParityTraceProvider{}
	require.Equal(t, "trace_block", provider.Method())
}

func TestParityTraceProvider_Params(t *testing.T) {
	provider := &ParityTraceProvider{}

	params := provider.Params("0x1a4")
	require.Len(t, params, 1)
	require.Equal(t, "0x1a4", params[0])
}

func TestParityTraceProvider_ParseTraces_SingleCall(t *testing.T) {
	raw := stdjson.RawMessage(`[
		{
			"action": {
				"from": "0x1111111111111111111111111111111111111111",
				"to": "0x2222222222222222222222222222222222222222",
				"gas": "0x5208",
				"value": "0x0",
				"input": "0x",
				"callType": "call"
			},
			"blockHash": "0xaaaa",
			"blockNumber": 100,
			"result": {
				"gasUsed": "0x5208",
				"output": "0x"
			},
			"subtraces": 0,
			"traceAddress": [],
			"transactionHash": "0xbbbb",
			"transactionPosition": 0,
			"type": "call"
		}
	]`)

	provider := &ParityTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)
	require.Len(t, traces, 1)

	tr := traces[0]
	require.Equal(t, "call", tr.Type)
	require.NotNil(t, tr.Action.From)
	require.NotNil(t, tr.Action.To)
	require.NotNil(t, tr.Action.Gas)
	require.NotNil(t, tr.Action.Value)
	require.NotNil(t, tr.Action.Input)
	require.NotNil(t, tr.Action.CallType)
	require.Equal(t, "call", *tr.Action.CallType)
	require.Equal(t, uint64(100), tr.BlockNumber)
	require.Equal(t, uint64(0), tr.Subtraces)
	require.Empty(t, tr.TraceAddress)
	require.NotNil(t, tr.TxHash)
	require.NotNil(t, tr.TxPosition)
	require.Equal(t, uint64(0), *tr.TxPosition)
}

func TestParityTraceProvider_ParseTraces_MultipleTraces(t *testing.T) {
	raw := stdjson.RawMessage(`[
		{
			"action": {
				"from": "0x1111111111111111111111111111111111111111",
				"to": "0x2222222222222222222222222222222222222222",
				"gas": "0xffff",
				"value": "0x100",
				"input": "0xabcd",
				"callType": "call"
			},
			"blockHash": "0xaaaa",
			"blockNumber": 200,
			"result": {
				"gasUsed": "0xaaaa",
				"output": "0x"
			},
			"subtraces": 1,
			"traceAddress": [],
			"transactionHash": "0xbbbb",
			"transactionPosition": 0,
			"type": "call"
		},
		{
			"action": {
				"from": "0x2222222222222222222222222222222222222222",
				"to": "0x3333333333333333333333333333333333333333",
				"gas": "0x1000",
				"value": "0x0",
				"input": "0x1234",
				"callType": "staticcall"
			},
			"blockHash": "0xaaaa",
			"blockNumber": 200,
			"result": {
				"gasUsed": "0x500",
				"output": "0x5678"
			},
			"subtraces": 0,
			"traceAddress": [0],
			"transactionHash": "0xbbbb",
			"transactionPosition": 0,
			"type": "call"
		}
	]`)

	provider := &ParityTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)
	require.Len(t, traces, 2)

	// Root trace
	require.Equal(t, "call", traces[0].Type)
	require.Equal(t, uint64(1), traces[0].Subtraces)
	require.Empty(t, traces[0].TraceAddress)

	// Child trace
	require.Equal(t, "call", traces[1].Type)
	require.Equal(t, "staticcall", *traces[1].Action.CallType)
	require.Equal(t, uint64(0), traces[1].Subtraces)
	require.Equal(t, []uint64{0}, traces[1].TraceAddress)
}

func TestParityTraceProvider_ParseTraces_CreateTrace(t *testing.T) {
	raw := stdjson.RawMessage(`[
		{
			"action": {
				"from": "0x1111111111111111111111111111111111111111",
				"gas": "0x10000",
				"value": "0x0",
				"init": "0x6080604052"
			},
			"blockHash": "0xaaaa",
			"blockNumber": 300,
			"result": {
				"gasUsed": "0x8000",
				"address": "0x9999999999999999999999999999999999999999",
				"code": "0x6080604052"
			},
			"subtraces": 0,
			"traceAddress": [],
			"transactionHash": "0xcccc",
			"transactionPosition": 0,
			"type": "create"
		}
	]`)

	provider := &ParityTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)
	require.Len(t, traces, 1)

	tr := traces[0]
	require.Equal(t, "create", tr.Type)
	require.NotNil(t, tr.Action.Init)
	require.Nil(t, tr.Action.To)
	require.Nil(t, tr.Action.CallType)
	require.NotNil(t, tr.Result.Address)
	require.NotNil(t, tr.Result.Code)
}

func TestParityTraceProvider_ParseTraces_RewardTrace(t *testing.T) {
	raw := stdjson.RawMessage(`[
		{
			"action": {
				"author": "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"rewardType": "block",
				"value": "0x1bc16d674ec80000"
			},
			"blockHash": "0xbbbb",
			"blockNumber": 400,
			"result": {},
			"subtraces": 0,
			"traceAddress": [],
			"type": "reward"
		}
	]`)

	provider := &ParityTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)
	require.Len(t, traces, 1)

	tr := traces[0]
	require.Equal(t, "reward", tr.Type)
	require.NotNil(t, tr.Action.Author)
	require.NotNil(t, tr.Action.RewardType)
	require.Equal(t, "block", *tr.Action.RewardType)
	require.NotNil(t, tr.Action.Value)
	require.Nil(t, tr.TxHash)
	require.Nil(t, tr.TxPosition)
}

func TestParityTraceProvider_ParseTraces_ErrorTrace(t *testing.T) {
	raw := stdjson.RawMessage(`[
		{
			"action": {
				"from": "0x1111111111111111111111111111111111111111",
				"to": "0x2222222222222222222222222222222222222222",
				"gas": "0x5208",
				"value": "0x0",
				"input": "0x",
				"callType": "call"
			},
			"blockHash": "0xaaaa",
			"blockNumber": 500,
			"error": "execution reverted",
			"subtraces": 0,
			"traceAddress": [],
			"transactionHash": "0xdddd",
			"transactionPosition": 0,
			"type": "call"
		}
	]`)

	provider := &ParityTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)
	require.Len(t, traces, 1)

	tr := traces[0]
	require.NotNil(t, tr.Error)
	require.Equal(t, "execution reverted", *tr.Error)
}

func TestParityTraceProvider_ParseTraces_SuicideTrace(t *testing.T) {
	raw := stdjson.RawMessage(`[
		{
			"action": {
				"address": "0x0020000000000000000000000000000000000000",
				"balance": "0x300",
				"refundAddress": "0x0000000000000999000000000000000000000000"
			},
			"blockHash": "0xaaaa",
			"blockNumber": 600,
			"result": null,
			"subtraces": 0,
			"traceAddress": [0],
			"transactionHash": "0xeeee",
			"transactionPosition": 0,
			"type": "suicide"
		}
	]`)

	provider := &ParityTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)
	require.Len(t, traces, 1)

	tr := traces[0]
	require.Equal(t, "suicide", tr.Type)
	// Parity uses address/refundAddress/balance (not from/to/value)
	require.NotNil(t, tr.Action.Address)
	require.Equal(t, mustHex("0x0020000000000000000000000000000000000000"), *tr.Action.Address)
	require.NotNil(t, tr.Action.RefundAddress)
	require.Equal(t, mustHex("0x0000000000000999000000000000000000000000"), *tr.Action.RefundAddress)
	require.NotNil(t, tr.Action.Balance)
	require.Equal(t, mustHex("0x300"), *tr.Action.Balance)
	// from/to/value should be nil for Parity selfdestruct
	require.Nil(t, tr.Action.From)
	require.Nil(t, tr.Action.To)
	require.Nil(t, tr.Action.Value)
	require.Nil(t, tr.Action.Gas)
	require.Nil(t, tr.Action.Input)
	require.Nil(t, tr.Action.CallType)
}

func TestParityTraceProvider_ParseTraces_EmptyArray(t *testing.T) {
	raw := stdjson.RawMessage(`[]`)

	provider := &ParityTraceProvider{}
	traces, err := provider.ParseTraces(raw)
	require.NoError(t, err)
	require.Empty(t, traces)
}

func TestParityTraceProvider_ParseTraces_InvalidJSON(t *testing.T) {
	raw := stdjson.RawMessage(`invalid json`)

	provider := &ParityTraceProvider{}
	_, err := provider.ParseTraces(raw)
	require.Error(t, err)
}

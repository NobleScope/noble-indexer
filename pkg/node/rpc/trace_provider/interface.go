package trace_provider

import (
	stdjson "encoding/json"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
)

// ITraceProvider abstracts the trace fetching strategy.
// Different networks use different RPC methods and response formats for traces.
type ITraceProvider interface {
	// Method returns the RPC method name (e.g. "trace_block", "debug_traceBlockByNumber").
	Method() string

	// Params returns the RPC call parameters for the given block level hex string.
	Params(hexLevel string) []any

	// ParseTraces parses the raw JSON-RPC response into list of traces.
	ParseTraces(raw stdjson.RawMessage) ([]pkgTypes.Trace, error)
}

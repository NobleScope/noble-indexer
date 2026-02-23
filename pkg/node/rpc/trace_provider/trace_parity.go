package trace_provider

import (
	stdjson "encoding/json"
	"github.com/pkg/errors"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
)

type ParityTraceProvider struct{}

func (p *ParityTraceProvider) Method() string {
	return "trace_block"
}

func (p *ParityTraceProvider) Params(hexLevel string) []any {
	return []any{hexLevel}
}

func (p *ParityTraceProvider) ParseTraces(raw stdjson.RawMessage) ([]pkgTypes.Trace, error) {
	var traces []pkgTypes.Trace
	if err := json.Unmarshal(raw, &traces); err != nil {
		return nil, errors.Wrap(err, "unmarshal parity trace results")
	}
	return traces, nil
}

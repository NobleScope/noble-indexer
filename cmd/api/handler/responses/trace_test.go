package responses

import (
	"testing"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func newTrace(traceAddr []uint64, subtraces uint64, traceType types.TraceType) *storage.Trace {
	return &storage.Trace{
		TraceAddress: traceAddr,
		Subtraces:    subtraces,
		Type:         traceType,
		GasLimit:     decimal.NewFromInt(21000),
		GasUsed:      decimal.NewFromInt(21000),
	}
}

func TestBuildTraceTree_NilOnEmpty(t *testing.T) {
	root, err := BuildTraceTree(nil)
	require.NoError(t, err)
	require.Nil(t, root)
}

func TestBuildTraceTree_SingleRoot(t *testing.T) {
	traces := []*storage.Trace{
		newTrace([]uint64{}, 0, "call"),
	}

	root, err := BuildTraceTree(traces)
	require.NoError(t, err)
	require.NotNil(t, root)
	require.NotNil(t, root.Trace)
	require.Equal(t, "call", root.Type)
	require.Empty(t, root.Children)
}

func TestBuildTraceTree_TwoLevels(t *testing.T) {
	// root -> [child0, child1]
	traces := []*storage.Trace{
		newTrace([]uint64{}, 2, "call"),
		newTrace([]uint64{0}, 0, "staticcall"),
		newTrace([]uint64{1}, 0, "delegatecall"),
	}

	root, err := BuildTraceTree(traces)
	require.NoError(t, err)
	require.NotNil(t, root)
	require.Len(t, root.Children, 2)
	require.Equal(t, "staticcall", root.Children[0].Type)
	require.Equal(t, "delegatecall", root.Children[1].Type)
}

func TestBuildTraceTree_ThreeLevels(t *testing.T) {
	// root -> [child0] -> [grandchild0]
	traces := []*storage.Trace{
		newTrace([]uint64{}, 1, "call"),
		newTrace([]uint64{0}, 1, "call"),
		newTrace([]uint64{0, 0}, 0, "staticcall"),
	}

	root, err := BuildTraceTree(traces)
	require.NoError(t, err)
	require.NotNil(t, root)
	require.Len(t, root.Children, 1)
	require.Len(t, root.Children[0].Children, 1)
	require.Equal(t, "staticcall", root.Children[0].Children[0].Type)
}

func TestBuildTraceTree_ErrorNoRoot(t *testing.T) {
	// first-level trace without a root — should fail
	traces := []*storage.Trace{
		newTrace([]uint64{0}, 0, "call"),
	}

	_, err := BuildTraceTree(traces)
	require.ErrorIs(t, err, errInvalidTraceAddress)
}

func TestBuildTraceTree_ErrorOutOfRange(t *testing.T) {
	// root has 1 subtrace but child references index 5
	traces := []*storage.Trace{
		newTrace([]uint64{}, 1, "call"),
		newTrace([]uint64{5}, 0, "call"),
	}

	_, err := BuildTraceTree(traces)
	require.ErrorIs(t, err, errInvalidTraceAddress)
}

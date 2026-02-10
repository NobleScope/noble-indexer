package postgres

import (
	"context"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
)

// TestTraceFilterBasic tests basic Filter functionality
func (s *StorageTestSuite) TestTraceFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 10)

	// Check default sort by time, id ascending
	for i := 1; i < len(traces); i++ {
		s.Require().True(
			traces[i-1].Time.Before(traces[i].Time) ||
				(traces[i-1].Time.Equal(traces[i].Time) && traces[i-1].Id <= traces[i].Id),
		)
	}
}

// TestTraceFilterByTxId tests filtering by tx_id
func (s *StorageTestSuite) TestTraceFilterByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(1)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		TxId:   &txId,
		Limit:  15,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 5) // tx 1 has 5 traces

	// All should have tx_id = 1
	for _, trace := range traces {
		s.Require().NotNil(trace.TxId)
		s.Require().EqualValues(1, *trace.TxId)
	}
}

// TestTraceFilterByAddressFromId tests filtering by from_address_id
func (s *StorageTestSuite) TestTraceFilterByAddressFromId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		AddressFromId: &addressId,
		Limit:         15,
		Offset:        0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 3) // address 1 sent 3 traces

	// All should have from_address_id = 1
	for _, trace := range traces {
		s.Require().NotNil(trace.From)
		s.Require().EqualValues(1, *trace.From)
	}
}

// TestTraceFilterByAddressToId tests filtering by to_address_id
func (s *StorageTestSuite) TestTraceFilterByAddressToId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(2)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		AddressToId: &addressId,
		Limit:       15,
		Offset:      0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 1) // address 2 received 1 trace

	// All should have to_address_id = 2
	for _, trace := range traces {
		s.Require().NotNil(trace.To)
		s.Require().EqualValues(2, *trace.To)
	}
}

// TestTraceFilterByContractId tests filtering by contract_id
func (s *StorageTestSuite) TestTraceFilterByContractId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(3)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		ContractId: &contractId,
		Limit:      15,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 3) // contract 3 has 3 traces

	// All should have contract_id = 3
	for _, trace := range traces {
		s.Require().NotNil(trace.ContractId)
		s.Require().EqualValues(3, *trace.ContractId)
	}
}

// TestTraceFilterByHeight tests filtering by height
func (s *StorageTestSuite) TestTraceFilterByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(100)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Height: &height,
		Limit:  15,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 2) // height 100 has 2 traces

	// All should have height = 100
	for _, trace := range traces {
		s.Require().EqualValues(100, trace.Height)
	}
}

// TestTraceFilterByType tests filtering by trace type
func (s *StorageTestSuite) TestTraceFilterByType() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Type:   []types.TraceType{"call"},
		Limit:  15,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 11) // 11 call traces

	// All should have type = call
	for _, trace := range traces {
		s.Require().Equal(types.TraceType("call"), trace.Type)
	}
}

// TestTraceFilterByMultipleTypes tests filtering by multiple trace types
func (s *StorageTestSuite) TestTraceFilterByMultipleTypes() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Type:   []types.TraceType{"create", "reward"},
		Limit:  15,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 3) // 2 create + 1 reward

	// All should have type in (create, reward)
	for _, trace := range traces {
		s.Require().True(trace.Type == types.TraceType("create") || trace.Type == types.TraceType("reward"))
	}
}

// TestTraceFilterCombined tests filtering by multiple parameters
func (s *StorageTestSuite) TestTraceFilterCombined() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(2)
	addressFromId := uint64(3)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		TxId:          &txId,
		AddressFromId: &addressFromId,
		Limit:         15,
		Offset:        0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 1) // only one trace matches

	s.Require().NotNil(traces[0].TxId)
	s.Require().EqualValues(2, *traces[0].TxId)
	s.Require().NotNil(traces[0].From)
	s.Require().EqualValues(3, *traces[0].From)
}

// TestTraceFilterWithLimit tests Filter with limit
func (s *StorageTestSuite) TestTraceFilterWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  3,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 3)
}

// TestTraceFilterWithOffset tests Filter with offset
func (s *StorageTestSuite) TestTraceFilterWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  3,
		Offset: 2,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 3)
}

// TestTraceFilterWithLimitOffset tests Filter with both limit and offset
func (s *StorageTestSuite) TestTraceFilterWithLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	allTraces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  15,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(allTraces, 15)

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  2,
		Offset: 3,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 2)

	// Verify pagination consistency
	s.Require().EqualValues(allTraces[3].Id, traces[0].Id)
	s.Require().EqualValues(allTraces[4].Id, traces[1].Id)
}

// TestTraceFilterSortDesc tests Filter with descending sort
func (s *StorageTestSuite) TestTraceFilterSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   "desc",
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 5)

	// Check descending order by time
	for i := 1; i < len(traces); i++ {
		s.Require().True(
			traces[i-1].Time.After(traces[i].Time) ||
				(traces[i-1].Time.Equal(traces[i].Time) && traces[i-1].Id >= traces[i].Id),
		)
	}
}

// TestTraceFilterSortAsc tests Filter with explicit ascending sort
func (s *StorageTestSuite) TestTraceFilterSortAsc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   "asc",
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 5)

	// Check ascending order by time
	for i := 1; i < len(traces); i++ {
		s.Require().True(
			traces[i-1].Time.Before(traces[i].Time) ||
				(traces[i-1].Time.Equal(traces[i].Time) && traces[i-1].Id <= traces[i].Id),
		)
	}
}

// TestTraceFilterInvalidSort tests Filter with invalid sort value (should default to asc)
func (s *StorageTestSuite) TestTraceFilterInvalidSort() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  3,
		Offset: 0,
		Sort:   "invalid",
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 3)

	// Should default to ascending order by time
	for i := 1; i < len(traces); i++ {
		s.Require().True(
			traces[i-1].Time.Before(traces[i].Time) ||
				(traces[i-1].Time.Equal(traces[i].Time) && traces[i-1].Id <= traces[i].Id),
		)
	}
}

// TestTraceFilterNoResults tests Filter with no matching results
func (s *StorageTestSuite) TestTraceFilterNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(999)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		TxId:   &txId,
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 0)
}

// TestTraceFilterOffsetExceedsTotal tests Filter with offset exceeding total
func (s *StorageTestSuite) TestTraceFilterOffsetExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  10,
		Offset: 100,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 0)
}

// TestTraceFilterJoinFields tests that JOIN fields are populated correctly
func (s *StorageTestSuite) TestTraceFilterJoinFields() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  1,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 1)

	// Check JOIN fields are populated
	s.Require().NotNil(traces[0].Tx)
	s.Require().NotEmpty(traces[0].Tx.Hash)

	if traces[0].From != nil {
		s.Require().NotNil(traces[0].FromAddress)
		s.Require().NotEmpty(traces[0].FromAddress.Hash)
	}

	if traces[0].To != nil {
		s.Require().NotNil(traces[0].ToAddress)
		s.Require().NotEmpty(traces[0].ToAddress.Hash)
	}
}

// TestTraceFilterZeroLimit tests Filter with zero limit (should use default 10)
func (s *StorageTestSuite) TestTraceFilterZeroLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  0,
		Offset: 0,
	})
	s.Require().NoError(err)
	// Should use default limit 10
	s.Require().Len(traces, 10)
}

// TestTraceFilterLimitExceedsMax tests Filter with limit > 100 (should use default 10)
func (s *StorageTestSuite) TestTraceFilterLimitExceedsMax() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  200,
		Offset: 0,
	})
	s.Require().NoError(err)
	// Should be limited to default 10
	s.Require().Len(traces, 10)
}

// TestTraceFilterNegativeLimit tests Filter with negative limit (should use default 10)
func (s *StorageTestSuite) TestTraceFilterNegativeLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		Limit:  -5,
		Offset: 0,
	})
	s.Require().NoError(err)
	// Should use default limit 10
	s.Require().Len(traces, 10)
}

// TestTraceFilterDifferentTypes tests filtering different trace types
func (s *StorageTestSuite) TestTraceFilterDifferentTypes() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		traceType     types.TraceType
		expectedCount int
	}{
		{"call", 11},
		{"create", 2},
		{"reward", 1},
		{"suicide", 1},
	}

	for _, tc := range testCases {
		traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
			Type:   []types.TraceType{tc.traceType},
			Limit:  15,
			Offset: 0,
		})
		s.Require().NoError(err)
		s.Require().Len(traces, tc.expectedCount, "trace type: %s", tc.traceType)

		// Verify all have correct type
		for _, trace := range traces {
			s.Require().Equal(tc.traceType, trace.Type)
		}
	}
}

// TestTraceFilterMultipleTxs tests filtering across multiple transactions
func (s *StorageTestSuite) TestTraceFilterMultipleTxs() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		txId          uint64
		expectedCount int
	}{
		{1, 5},
		{2, 6},
		{3, 4},
	}

	for _, tc := range testCases {
		traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
			TxId:   &tc.txId,
			Limit:  15,
			Offset: 0,
		})
		s.Require().NoError(err)
		s.Require().Len(traces, tc.expectedCount, "tx_id: %d", tc.txId)

		// Verify all have correct tx_id
		for _, trace := range traces {
			s.Require().NotNil(trace.TxId)
			s.Require().EqualValues(tc.txId, *trace.TxId)
		}
	}
}

// TestTraceFilterMultipleContracts tests filtering across multiple contracts
func (s *StorageTestSuite) TestTraceFilterMultipleContracts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		contractId    uint64
		expectedCount int
	}{
		{3, 3},
		{4, 3},
		{5, 2},
		{6, 1},
		{7, 1},
	}

	for _, tc := range testCases {
		traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
			ContractId: &tc.contractId,
			Limit:      15,
			Offset:     0,
		})
		s.Require().NoError(err)
		s.Require().Len(traces, tc.expectedCount, "contract_id: %d", tc.contractId)

		// Verify all have correct contract_id
		for _, trace := range traces {
			s.Require().NotNil(trace.ContractId)
			s.Require().EqualValues(tc.contractId, *trace.ContractId)
		}
	}
}

// TestTraceFilterMultipleHeights tests filtering across multiple heights
func (s *StorageTestSuite) TestTraceFilterMultipleHeights() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		height        uint64
		expectedCount int
	}{
		{100, 2},
		{150, 2},
		{200, 2},
		{250, 2},
		{300, 2},
		{350, 2},
		{400, 2},
		{450, 1},
	}

	for _, tc := range testCases {
		traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
			Height: &tc.height,
			Limit:  15,
			Offset: 0,
		})
		s.Require().NoError(err)
		s.Require().Len(traces, tc.expectedCount, "height: %d", tc.height)

		// Verify all have correct height
		for _, trace := range traces {
			s.Require().EqualValues(tc.height, trace.Height)
		}
	}
}

func (s *StorageTestSuite) TestTraceFilterByAddressId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(3)
	traces, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{
		AddressId: &addressId,
		Limit:     10,
		Offset:    0,
	})
	s.Require().NoError(err)
	s.Require().Len(traces, 3)

	// All should have from_address_id = 3 or to_address_id = 3
	for _, trace := range traces {
		s.Require().True(
			(trace.From != nil && *trace.From == addressId) ||
				(trace.To != nil && *trace.To == addressId),
		)
	}
}

func (s *StorageTestSuite) TestTraceByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(2)
	traces, err := s.storage.Trace.ByTxId(ctx, txId)
	s.Require().NoError(err)
	s.Require().Len(traces, 6) // tx 2 has 6 traces

	// All should have tx_id = 2
	for _, trace := range traces {
		s.Require().NotNil(trace.TxId)
		s.Require().EqualValues(2, *trace.TxId)
	}
}

package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

// TestTxFilterBasic tests basic Filter functionality
func (s *StorageTestSuite) TestTxFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 10)

	// Check first tx
	s.Require().EqualValues(1, txs[0].Id)
	s.Require().EqualValues(100, txs[0].Height)
	s.Require().EqualValues(types.TxTypeDynamicFee, txs[0].Type)
	s.Require().EqualValues(types.TxStatusSuccess, txs[0].Status)
	s.Require().EqualValues(1, txs[0].FromAddressId)
	s.Require().NotNil(txs[0].ToAddressId)
	s.Require().EqualValues(2, *txs[0].ToAddressId)

	// Check that JOIN fields are populated
	s.Require().NotNil(txs[0].FromAddress.Hash)
	s.Require().NotNil(txs[0].ToAddress)
	s.Require().NotNil(txs[0].ToAddress.Hash)
}

// TestTxFilterByHeight tests filtering by height
func (s *StorageTestSuite) TestTxFilterByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		height        uint64
		expectedCount int
	}{
		{100, 4},
		{200, 5},
		{300, 6},
	}

	for _, tc := range testCases {
		txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
			Height: &tc.height,
			Limit:  15,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(txs, tc.expectedCount, "height: %d", tc.height)

		// Verify all have correct height
		for _, tx := range txs {
			s.Require().EqualValues(tc.height, tx.Height)
		}
	}
}

// TestTxFilterByStatus tests filtering by status
func (s *StorageTestSuite) TestTxFilterByStatus() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		status        types.TxStatus
		expectedCount int
	}{
		{types.TxStatusSuccess, 13},
		{types.TxStatusRevert, 2},
	}

	for _, tc := range testCases {
		txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
			Status: []types.TxStatus{tc.status},
			Limit:  15,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(txs, tc.expectedCount, "status: %s", tc.status)

		// Verify all have correct status
		for _, tx := range txs {
			s.Require().Equal(tc.status, tx.Status)
		}
	}
}

// TestTxFilterByType tests filtering by type
func (s *StorageTestSuite) TestTxFilterByType() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		txType        types.TxType
		expectedCount int
	}{
		{types.TxTypeLegacy, 3},
		{types.TxTypeDynamicFee, 11},
		{types.TxTypeUnknown, 1},
	}

	for _, tc := range testCases {
		txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
			Type:   []types.TxType{tc.txType},
			Limit:  15,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(txs, tc.expectedCount, "type: %s", tc.txType)

		// Verify all have correct type
		for _, tx := range txs {
			s.Require().Equal(tc.txType, tx.Type)
		}
	}
}

// TestTxFilterByMultipleTypes tests filtering by multiple types
func (s *StorageTestSuite) TestTxFilterByMultipleTypes() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Type:   []types.TxType{types.TxTypeLegacy, types.TxTypeUnknown},
		Limit:  15,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 4)

	// Verify all are either Legacy or Unknown
	for _, tx := range txs {
		s.Require().True(
			tx.Type == types.TxTypeLegacy || tx.Type == types.TxTypeUnknown,
		)
	}
}

// TestTxFilterByAddressFromId tests filtering by from_address_id
func (s *StorageTestSuite) TestTxFilterByAddressFromId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		AddressFromId: &addressId,
		Limit:         15,
		Offset:        0,
		Sort:          sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 2)

	// Verify all have from_address_id = 1
	for _, tx := range txs {
		s.Require().EqualValues(1, tx.FromAddressId)
	}
}

// TestTxFilterByAddressToId tests filtering by to_address_id
func (s *StorageTestSuite) TestTxFilterByAddressToId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(2)
	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		AddressToId: &addressId,
		Limit:       15,
		Offset:      0,
		Sort:        sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 2)

	// Verify all have to_address_id = 2
	for _, tx := range txs {
		s.Require().NotNil(tx.ToAddressId)
		s.Require().EqualValues(2, *tx.ToAddressId)
	}
}

// TestTxFilterByTimeRange tests filtering by time range
func (s *StorageTestSuite) TestTxFilterByTimeRange() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:00Z")

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
		Limit:    15,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 5)

	// Verify all are within time range
	for _, tx := range txs {
		s.Require().True(tx.Time.Equal(timeFrom) || tx.Time.After(timeFrom))
		s.Require().True(tx.Time.Before(timeTo))
	}
}

// TestTxFilterWithLimit tests Filter with limit
func (s *StorageTestSuite) TestTxFilterWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  3,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 3)
}

// TestTxFilterWithOffset tests Filter with offset
func (s *StorageTestSuite) TestTxFilterWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  3,
		Offset: 2,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 3)
}

// TestTxFilterWithLimitOffset tests Filter with both limit and offset
func (s *StorageTestSuite) TestTxFilterWithLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Get all txs first
	allTxs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  15,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(allTxs, 15)

	// Get with offset
	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  2,
		Offset: 3,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 2)

	// Verify pagination consistency
	s.Require().EqualValues(allTxs[3].Id, txs[0].Id)
	s.Require().EqualValues(allTxs[4].Id, txs[1].Id)
}

// TestTxFilterSortDesc tests Filter with descending sort
func (s *StorageTestSuite) TestTxFilterSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 5)

	// Check descending order by time then id
	for i := 1; i < len(txs); i++ {
		if txs[i-1].Time.Equal(txs[i].Time) {
			s.Require().True(txs[i-1].Id >= txs[i].Id)
		} else {
			s.Require().True(txs[i-1].Time.After(txs[i].Time))
		}
	}
}

// TestTxFilterNoResults tests Filter with no matching results
func (s *StorageTestSuite) TestTxFilterNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(999)
	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Height: &height,
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 0)
}

// TestTxFilterOffsetExceedsTotal tests Filter with offset exceeding total
func (s *StorageTestSuite) TestTxFilterOffsetExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  10,
		Offset: 100,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 0)
}

// TestTxFilterZeroLimit tests Filter with zero limit (should use default 10)
func (s *StorageTestSuite) TestTxFilterZeroLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  0,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 10)
}

// TestTxFilterLimitExceedsMax tests Filter with limit > 100 (should use default 10)
func (s *StorageTestSuite) TestTxFilterLimitExceedsMax() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Limit:  200,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 10)
}

// TestTxFilterCombinedFilters tests filtering with multiple filters
func (s *StorageTestSuite) TestTxFilterCombinedFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(100)
	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Height: &height,
		Type:   []types.TxType{types.TxTypeDynamicFee},
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 3)

	// Verify all match both filters
	for _, tx := range txs {
		s.Require().EqualValues(100, tx.Height)
		s.Require().Equal(types.TxTypeDynamicFee, tx.Type)
	}
}

// TestTxFilterMultipleStatuses tests filtering by multiple statuses
func (s *StorageTestSuite) TestTxFilterMultipleStatuses() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		Status: []types.TxStatus{types.TxStatusSuccess, types.TxStatusRevert},
		Limit:  15,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 15)

	// Verify all are either Success or Revert
	for _, tx := range txs {
		s.Require().True(
			tx.Status == types.TxStatusSuccess || tx.Status == types.TxStatusRevert,
		)
	}
}

// TestTxByHeight tests ByHeight method
func (s *StorageTestSuite) TestTxByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.ByHeight(ctx, 100, 10, 0, sdk.SortOrderAsc)
	s.Require().NoError(err)
	s.Require().Len(txs, 4)

	// Verify all have height = 100
	for _, tx := range txs {
		s.Require().EqualValues(100, tx.Height)
	}
}

// TestTxByHeightWithPagination tests ByHeight with pagination
func (s *StorageTestSuite) TestTxByHeightWithPagination() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Get first page
	txsPage1, err := s.storage.Tx.ByHeight(ctx, 100, 2, 0, sdk.SortOrderAsc)
	s.Require().NoError(err)
	s.Require().Len(txsPage1, 2)

	// Get second page
	txsPage2, err := s.storage.Tx.ByHeight(ctx, 100, 2, 2, sdk.SortOrderAsc)
	s.Require().NoError(err)
	s.Require().Len(txsPage2, 2)

	// Verify they are different
	s.Require().NotEqual(txsPage1[0].Id, txsPage2[0].Id)
}

// TestTxByHash tests ByHash method
func (s *StorageTestSuite) TestTxByHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("90f5df4e03620cc55d3ea295bf8826f84465065340cb6d0d095166dd2465f283")
	tx, err := s.storage.Tx.ByHash(ctx, hash)
	s.Require().NoError(err)
	s.Require().EqualValues(1, tx.Id)
	s.Require().EqualValues(hash, tx.Hash)
}

// TestTxFilterDifferentHeights tests filtering by different heights
func (s *StorageTestSuite) TestTxFilterDifferentHeights() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		height        uint64
		expectedCount int
	}{
		{100, 4},
		{200, 5},
		{300, 6},
	}

	for _, tc := range testCases {
		height := tc.height
		txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
			Height: &height,
			Limit:  15,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(txs, tc.expectedCount, "height: %d", tc.height)

		// Verify all have correct height
		for _, tx := range txs {
			s.Require().EqualValues(tc.height, tx.Height)
		}
	}
}

// TestTxByHeightSortDesc tests ByHeight with descending sort
func (s *StorageTestSuite) TestTxByHeightSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.ByHeight(ctx, 100, 10, 0, sdk.SortOrderDesc)
	s.Require().NoError(err)
	s.Require().Len(txs, 4)

	// Verify descending order by id
	for i := 1; i < len(txs); i++ {
		s.Require().True(txs[i-1].Id >= txs[i].Id)
	}
}

// TestTxByHeightNonExistent tests ByHeight with non-existent height
func (s *StorageTestSuite) TestTxByHeightNonExistent() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txs, err := s.storage.Tx.ByHeight(ctx, 999, 10, 0, sdk.SortOrderAsc)
	s.Require().NoError(err)
	s.Require().Len(txs, 0)
}

// TestTxByHashNonExistent tests ByHash with non-existent hash
func (s *StorageTestSuite) TestTxByHashNonExistent() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("0000000000000000000000000000000000000000000000000000000000000000")
	_, err := s.storage.Tx.ByHash(ctx, hash)
	s.Require().Error(err)
}

// TestTxFilterByContractId tests filtering by contract_id (from or to address)
func (s *StorageTestSuite) TestTxFilterByContractId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(1)
	txs, err := s.storage.Tx.Filter(ctx, storage.TxListFilter{
		ContractId: &contractId,
		Limit:      15,
		Offset:     0,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(txs, 3) // txs where address 1 is from (2 txs) or to (1 tx)

	// Verify all have address 1 as either from or to
	for _, tx := range txs {
		hasAddress1 := tx.FromAddressId == 1 ||
			(tx.ToAddressId != nil && *tx.ToAddressId == 1)
		s.Require().True(hasAddress1, "tx %d should have address 1 as from or to", tx.Id)
	}
}

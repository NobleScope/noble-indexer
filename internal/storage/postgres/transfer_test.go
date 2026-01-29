package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
)

// TestTransferFilterBasic tests basic Filter functionality
func (s *StorageTestSuite) TestTransferFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 10)

	// Check first transfer
	s.Require().EqualValues(1, transfers[0].Id)
	s.Require().EqualValues(100, transfers[0].Height)
	s.Require().EqualValues(types.TransferType("transfer"), transfers[0].Type)
	s.Require().NotNil(transfers[0].FromAddressId)
	s.Require().EqualValues(1, *transfers[0].FromAddressId)
	s.Require().NotNil(transfers[0].ToAddressId)
	s.Require().EqualValues(2, *transfers[0].ToAddressId)
	s.Require().EqualValues(1, transfers[0].TxID)
	s.Require().True(transfers[0].TokenID.Equal(decimal.NewFromInt(0)))
	s.Require().True(transfers[0].Amount.Equal(decimal.NewFromInt(1000000000000000000)))

	// Check that JOIN fields are populated
	s.Require().NotNil(transfers[0].Tx.Hash)
	s.Require().NotNil(transfers[0].FromAddress)
	s.Require().NotNil(transfers[0].FromAddress.Hash)
	s.Require().NotNil(transfers[0].ToAddress)
	s.Require().NotNil(transfers[0].ToAddress.Hash)
	s.Require().NotNil(transfers[0].Contract.Address)
	s.Require().NotNil(transfers[0].Contract.Address.Hash)
	s.Require().NotNil(transfers[0].Token)
	s.Require().NotNil(transfers[0].Token.Name)
	s.Require().NotNil(transfers[0].Token.Symbol)
	s.Require().NotNil(transfers[0].Token.Decimals)
	s.Require().NotNil(transfers[0].Token.Type)
	s.Require().NotNil(transfers[0].Token.Supply)
	s.Require().NotNil(transfers[0].Token.TransfersCount)
}

// TestTransferFilterByTxId tests filtering by tx_id
func (s *StorageTestSuite) TestTransferFilterByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		txId          uint64
		expectedCount int
	}{
		{1, 5},
		{2, 5},
		{3, 5},
	}

	for _, tc := range testCases {
		transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
			TxId:   &tc.txId,
			Limit:  15,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(transfers, tc.expectedCount, "tx_id: %d", tc.txId)

		// Verify all have correct tx_id
		for _, transfer := range transfers {
			s.Require().EqualValues(tc.txId, transfer.TxID)
		}
	}
}

// TestTransferFilterByAddressFromId tests filtering by from_address_id
func (s *StorageTestSuite) TestTransferFilterByAddressFromId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		AddressFromId: &addressId,
		Limit:         15,
		Offset:        0,
		Sort:          sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 3)

	// Verify all have from_address_id = 1
	for _, transfer := range transfers {
		s.Require().NotNil(transfer.FromAddressId)
		s.Require().EqualValues(1, *transfer.FromAddressId)
	}
}

// TestTransferFilterByAddressToId tests filtering by to_address_id
func (s *StorageTestSuite) TestTransferFilterByAddressToId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(3)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		AddressToId: &addressId,
		Limit:       15,
		Offset:      0,
		Sort:        sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 2)

	// Verify all have to_address_id = 3
	for _, transfer := range transfers {
		s.Require().NotNil(transfer.ToAddressId)
		s.Require().EqualValues(3, *transfer.ToAddressId)
	}
}

// TestTransferFilterByContractId tests filtering by contract_id
func (s *StorageTestSuite) TestTransferFilterByContractId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(3)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		ContractId: &contractId,
		Limit:      15,
		Offset:     0,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 5)

	// Verify all have contract_id = 3
	for _, transfer := range transfers {
		s.Require().EqualValues(3, transfer.ContractId)
	}
}

// TestTransferFilterByHeight tests filtering by height
func (s *StorageTestSuite) TestTransferFilterByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(200)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Height: &height,
		Limit:  15,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 5)

	// Verify all have height = 200
	for _, transfer := range transfers {
		s.Require().EqualValues(200, transfer.Height)
	}
}

// TestTransferFilterByTokenId tests filtering by token_id
func (s *StorageTestSuite) TestTransferFilterByTokenId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokenId := decimal.NewFromInt(0)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		TokenId: &tokenId,
		Limit:   15,
		Offset:  0,
		Sort:    sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 8)

	// Verify all have token_id = 0
	for _, transfer := range transfers {
		s.Require().True(transfer.TokenID.Equal(decimal.NewFromInt(0)))
	}
}

// TestTransferFilterByType tests filtering by transfer type
func (s *StorageTestSuite) TestTransferFilterByType() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		transferType  types.TransferType
		expectedCount int
	}{
		{types.TransferType("transfer"), 9},
		{types.TransferType("burn"), 3},
		{types.TransferType("mint"), 3},
	}

	for _, tc := range testCases {
		transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
			Type:   []types.TransferType{tc.transferType},
			Limit:  15,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(transfers, tc.expectedCount, "type: %s", tc.transferType)

		// Verify all have correct type
		for _, transfer := range transfers {
			s.Require().Equal(tc.transferType, transfer.Type)
		}
	}
}

// TestTransferFilterByMultipleTypes tests filtering by multiple types
func (s *StorageTestSuite) TestTransferFilterByMultipleTypes() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Type:   []types.TransferType{types.TransferType("burn"), types.TransferType("mint")},
		Limit:  15,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 6)

	// Verify all are either burn or mint
	for _, transfer := range transfers {
		s.Require().True(
			transfer.Type == types.TransferType("burn") || transfer.Type == types.TransferType("mint"),
		)
	}
}

// TestTransferFilterByTimeRange tests filtering by time range
func (s *StorageTestSuite) TestTransferFilterByTimeRange() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:00Z")

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
		Limit:    15,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 5)

	// Verify all are within time range
	for _, transfer := range transfers {
		s.Require().True(transfer.Time.Equal(timeFrom) || transfer.Time.After(timeFrom))
		s.Require().True(transfer.Time.Before(timeTo))
	}
}

// TestTransferFilterWithLimit tests Filter with limit
func (s *StorageTestSuite) TestTransferFilterWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  3,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 3)
}

// TestTransferFilterWithOffset tests Filter with offset
func (s *StorageTestSuite) TestTransferFilterWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  3,
		Offset: 2,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 3)
}

// TestTransferFilterWithLimitOffset tests Filter with both limit and offset
func (s *StorageTestSuite) TestTransferFilterWithLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Get all transfers first
	allTransfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  15,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(allTransfers, 15)

	// Get with offset
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  2,
		Offset: 3,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 2)

	// Verify pagination consistency
	s.Require().EqualValues(allTransfers[3].Id, transfers[0].Id)
	s.Require().EqualValues(allTransfers[4].Id, transfers[1].Id)
}

// TestTransferFilterSortDesc tests Filter with descending sort
func (s *StorageTestSuite) TestTransferFilterSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 5)

	// Check descending order by time then id
	for i := 1; i < len(transfers); i++ {
		if transfers[i-1].Time.Equal(transfers[i].Time) {
			s.Require().True(transfers[i-1].Id >= transfers[i].Id)
		} else {
			s.Require().True(transfers[i-1].Time.After(transfers[i].Time))
		}
	}
}

// TestTransferFilterNoResults tests Filter with no matching results
func (s *StorageTestSuite) TestTransferFilterNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(999)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Height: &height,
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 0)
}

// TestTransferFilterOffsetExceedsTotal tests Filter with offset exceeding total
func (s *StorageTestSuite) TestTransferFilterOffsetExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  10,
		Offset: 100,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 0)
}

// TestTransferFilterZeroLimit tests Filter with zero limit (should use default 10)
func (s *StorageTestSuite) TestTransferFilterZeroLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  0,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 10)
}

// TestTransferFilterLimitExceedsMax tests Filter with limit > 100 (should use default 10)
func (s *StorageTestSuite) TestTransferFilterLimitExceedsMax() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Limit:  200,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 10)
}

// TestTransferFilterCombinedFilters tests filtering with multiple filters
func (s *StorageTestSuite) TestTransferFilterCombinedFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(3)
	height := uint64(100)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		ContractId: &contractId,
		Height:     &height,
		Limit:      10,
		Offset:     0,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 2)

	// Verify all match both filters
	for _, transfer := range transfers {
		s.Require().EqualValues(3, transfer.ContractId)
		s.Require().EqualValues(100, transfer.Height)
	}
}

// TestTransferFilterNullFromAddress tests filtering transfers with NULL from_address_id (mint)
func (s *StorageTestSuite) TestTransferFilterNullFromAddress() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Type:   []types.TransferType{types.TransferType("mint")},
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 3)

	// Verify all mint transfers have NULL from_address_id
	for _, transfer := range transfers {
		s.Require().Nil(transfer.FromAddressId)
		s.Require().Nil(transfer.FromAddress)
		s.Require().NotNil(transfer.ToAddressId)
	}
}

// TestTransferFilterNullToAddress tests filtering transfers with NULL to_address_id (burn)
func (s *StorageTestSuite) TestTransferFilterNullToAddress() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		Type:   []types.TransferType{types.TransferType("burn")},
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 3)

	// Verify all burn transfers have NULL to_address_id
	for _, transfer := range transfers {
		s.Require().Nil(transfer.ToAddressId)
		s.Require().Nil(transfer.ToAddress)
		s.Require().NotNil(transfer.FromAddressId)
	}
}

// TestTransferFilterDifferentTokenIds tests filtering by different token_ids
func (s *StorageTestSuite) TestTransferFilterDifferentTokenIds() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		tokenId       int64
		expectedCount int
	}{
		{0, 8},
		{1, 3},
		{2, 2},
		{5, 1},
	}

	for _, tc := range testCases {
		tokenId := decimal.NewFromInt(tc.tokenId)
		transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
			TokenId: &tokenId,
			Limit:   15,
			Offset:  0,
			Sort:    sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(transfers, tc.expectedCount, "token_id: %d", tc.tokenId)

		// Verify all have correct token_id
		for _, transfer := range transfers {
			s.Require().True(transfer.TokenID.Equal(decimal.NewFromInt(tc.tokenId)))
		}
	}
}

// TestTransferFilterAmounts tests that amount values are correct
func (s *StorageTestSuite) TestTransferFilterAmounts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(1)
	transfers, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{
		TxId:   &txId,
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(transfers, 5)

	// Check specific amount value for first transfer
	s.Require().True(transfers[0].Amount.Equal(decimal.NewFromInt(1000000000000000000)))
}

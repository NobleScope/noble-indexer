package postgres

import (
	"context"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
)

// TestBeaconWithdrawalFilterBasic tests basic Filter functionality
func (s *StorageTestSuite) TestBeaconWithdrawalFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 10)

	// Check first withdrawal
	s.Require().EqualValues(1, withdrawals[0].Id)
	s.Require().EqualValues(100, withdrawals[0].Height)
	s.Require().EqualValues(0, withdrawals[0].Index)
	s.Require().EqualValues(12345, withdrawals[0].ValidatorIndex)
	s.Require().EqualValues(1, withdrawals[0].AddressId)
	s.Require().True(withdrawals[0].Amount.Equal(decimal.RequireFromString("32000000000000000000")))

	// Check that JOIN fields are populated
	s.Require().NotNil(withdrawals[0].Address.Hash)
}

// TestBeaconWithdrawalFilterByHeight tests filtering by height
func (s *StorageTestSuite) TestBeaconWithdrawalFilterByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		height        pkgTypes.Level
		expectedCount int
	}{
		{100, 3},
		{200, 2},
		{300, 3},
		{400, 2},
	}

	for _, tc := range testCases {
		withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
			Height: &tc.height,
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(withdrawals, tc.expectedCount, "height: %d", tc.height)

		// Verify all have correct height
		for _, w := range withdrawals {
			s.Require().EqualValues(tc.height, w.Height)
		}
	}
}

// TestBeaconWithdrawalFilterByAddressId tests filtering by address_id
func (s *StorageTestSuite) TestBeaconWithdrawalFilterByAddressId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		addressId     uint64
		expectedCount int
	}{
		{1, 3},
		{2, 2},
		{3, 2},
		{4, 1},
		{5, 1},
		{6, 1},
	}

	for _, tc := range testCases {
		withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
			AddressId: &tc.addressId,
			Limit:     10,
			Offset:    0,
			Sort:      sdk.SortOrderAsc,
		})
		s.Require().NoError(err)
		s.Require().Len(withdrawals, tc.expectedCount, "address_id: %d", tc.addressId)

		// Verify all have correct address_id
		for _, w := range withdrawals {
			s.Require().EqualValues(tc.addressId, w.AddressId)
		}
	}
}

// TestBeaconWithdrawalFilterCombined tests filtering with multiple filters
func (s *StorageTestSuite) TestBeaconWithdrawalFilterCombined() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := pkgTypes.Level(100)
	addressId := uint64(1)
	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Height:    &height,
		AddressId: &addressId,
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 2) // id 1 and 3 match both filters

	// Verify all match both filters
	for _, w := range withdrawals {
		s.Require().EqualValues(100, w.Height)
		s.Require().EqualValues(1, w.AddressId)
	}
}

// TestBeaconWithdrawalFilterWithLimit tests Filter with limit
func (s *StorageTestSuite) TestBeaconWithdrawalFilterWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  3,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 3)
}

// TestBeaconWithdrawalFilterWithOffset tests Filter with offset
func (s *StorageTestSuite) TestBeaconWithdrawalFilterWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  3,
		Offset: 2,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 3)
	s.Require().EqualValues(3, withdrawals[0].Id)
}

// TestBeaconWithdrawalFilterWithLimitOffset tests Filter with both limit and offset
func (s *StorageTestSuite) TestBeaconWithdrawalFilterWithLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Get all withdrawals first
	allWithdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(allWithdrawals, 10)

	// Get with offset
	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  2,
		Offset: 3,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 2)

	// Verify pagination consistency
	s.Require().EqualValues(allWithdrawals[3].Id, withdrawals[0].Id)
	s.Require().EqualValues(allWithdrawals[4].Id, withdrawals[1].Id)
}

// TestBeaconWithdrawalFilterSortDesc tests Filter with descending sort
func (s *StorageTestSuite) TestBeaconWithdrawalFilterSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 5)

	// Check descending order by time then id
	for i := 1; i < len(withdrawals); i++ {
		if withdrawals[i-1].Time.Equal(withdrawals[i].Time) {
			s.Require().True(withdrawals[i-1].Id >= withdrawals[i].Id)
		} else {
			s.Require().True(withdrawals[i-1].Time.After(withdrawals[i].Time))
		}
	}
}

// TestBeaconWithdrawalFilterNoResults tests Filter with no matching results
func (s *StorageTestSuite) TestBeaconWithdrawalFilterNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := pkgTypes.Level(999)
	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Height: &height,
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 0)
}

// TestBeaconWithdrawalFilterOffsetExceedsTotal tests Filter with offset exceeding total
func (s *StorageTestSuite) TestBeaconWithdrawalFilterOffsetExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  10,
		Offset: 100,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 0)
}

// TestBeaconWithdrawalFilterZeroLimit tests Filter with zero limit (should use default 10)
func (s *StorageTestSuite) TestBeaconWithdrawalFilterZeroLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  0,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 10)
}

// TestBeaconWithdrawalFilterLimitExceedsMax tests Filter with limit > 100 (should use default 10)
func (s *StorageTestSuite) TestBeaconWithdrawalFilterLimitExceedsMax() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  200,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 10)
}

// TestBeaconWithdrawalFilterJoinFields tests that JOIN fields are populated correctly
func (s *StorageTestSuite) TestBeaconWithdrawalFilterJoinFields() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Limit:  1,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1)

	// Check JOIN field is populated
	s.Require().NotNil(withdrawals[0].Address.Hash)
}

// TestBeaconWithdrawalFilterValidatorIndex tests that validator_index is correct
func (s *StorageTestSuite) TestBeaconWithdrawalFilterValidatorIndex() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := pkgTypes.Level(100)
	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Height: &height,
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 3)

	// Check validator indices for height 100
	s.Require().EqualValues(12345, withdrawals[0].ValidatorIndex)
	s.Require().EqualValues(12346, withdrawals[1].ValidatorIndex)
	s.Require().EqualValues(12347, withdrawals[2].ValidatorIndex)
}

// TestBeaconWithdrawalFilterAmounts tests that amounts are correct
func (s *StorageTestSuite) TestBeaconWithdrawalFilterAmounts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := pkgTypes.Level(100)
	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Height: &height,
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 3)

	// Check amounts
	s.Require().True(withdrawals[0].Amount.Equal(decimal.RequireFromString("32000000000000000000")))
	s.Require().True(withdrawals[1].Amount.Equal(decimal.RequireFromString("1500000000000000000")))
	s.Require().True(withdrawals[2].Amount.Equal(decimal.RequireFromString("500000000000000000")))
}

// TestBeaconWithdrawalFilterIndex tests that withdrawal index within block is correct
func (s *StorageTestSuite) TestBeaconWithdrawalFilterIndex() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := pkgTypes.Level(300)
	withdrawals, err := s.storage.BeaconWithdrawal.Filter(ctx, storage.BeaconWithdrawalListFilter{
		Height: &height,
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 3)

	// Check indices are sequential within block
	s.Require().EqualValues(0, withdrawals[0].Index)
	s.Require().EqualValues(1, withdrawals[1].Index)
	s.Require().EqualValues(2, withdrawals[2].Index)
}

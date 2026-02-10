package postgres

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/NobleScope/noble-indexer/internal/storage"
)

// TestTokenBalanceFilterBasic tests basic Filter functionality
func (s *StorageTestSuite) TestTokenBalanceFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 10)

	// Check default sort by balance ascending
	for i := 1; i < len(balances); i++ {
		s.Require().True(balances[i-1].Balance.LessThanOrEqual(balances[i].Balance))
	}
}

// TestTokenBalanceFilterByAddressId tests filtering by address_id
func (s *StorageTestSuite) TestTokenBalanceFilterByAddressId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId: &addressId,
		Limit:     10,
		Offset:    0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 5) // address 1 has 5 balances

	// All should have address_id = 1
	for _, balance := range balances {
		s.Require().EqualValues(1, balance.AddressID)
	}
}

// TestTokenBalanceFilterByContractId tests filtering by contract_id
func (s *StorageTestSuite) TestTokenBalanceFilterByContractId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(3)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		ContractId: &contractId,
		Limit:      15,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 10) // contract 3 has 10 balances

	// All should have contract_id = 3
	for _, balance := range balances {
		s.Require().EqualValues(3, balance.ContractID)
	}
}

// TestTokenBalanceFilterByTokenId tests filtering by token_id
func (s *StorageTestSuite) TestTokenBalanceFilterByTokenId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokenId := decimal.NewFromInt(0)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		TokenId: &tokenId,
		Limit:   15,
		Offset:  0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 8) // token_id 0 has 8 balances

	// All should have token_id = 0
	for _, balance := range balances {
		s.Require().True(balance.TokenID.Equal(decimal.NewFromInt(0)))
	}
}

// TestTokenBalanceFilterByAddressAndContract tests filtering by address_id and contract_id
func (s *StorageTestSuite) TestTokenBalanceFilterByAddressAndContract() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	contractId := uint64(3)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId:  &addressId,
		ContractId: &contractId,
		Limit:      10,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 3) // address 1 + contract 3 has 3 balances

	for _, balance := range balances {
		s.Require().EqualValues(1, balance.AddressID)
		s.Require().EqualValues(3, balance.ContractID)
	}
}

// TestTokenBalanceFilterByAddressAndToken tests filtering by address_id and token_id
func (s *StorageTestSuite) TestTokenBalanceFilterByAddressAndToken() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	tokenId := decimal.NewFromInt(0)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId: &addressId,
		TokenId:   &tokenId,
		Limit:     10,
		Offset:    0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 3) // address 1 + token_id 0 has 3 balances

	for _, balance := range balances {
		s.Require().EqualValues(1, balance.AddressID)
		s.Require().True(balance.TokenID.Equal(decimal.NewFromInt(0)))
	}
}

// TestTokenBalanceFilterByContractAndToken tests filtering by contract_id and token_id
func (s *StorageTestSuite) TestTokenBalanceFilterByContractAndToken() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(3)
	tokenId := decimal.NewFromInt(0)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		ContractId: &contractId,
		TokenId:    &tokenId,
		Limit:      10,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 3) // contract 3 + token_id 0 has 3 balances

	for _, balance := range balances {
		s.Require().EqualValues(3, balance.ContractID)
		s.Require().True(balance.TokenID.Equal(decimal.NewFromInt(0)))
	}
}

// TestTokenBalanceFilterAllFilters tests filtering by all three filters
func (s *StorageTestSuite) TestTokenBalanceFilterAllFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	contractId := uint64(3)
	tokenId := decimal.NewFromInt(0)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId:  &addressId,
		ContractId: &contractId,
		TokenId:    &tokenId,
		Limit:      10,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 1) // exactly one balance matches all filters

	s.Require().EqualValues(1, balances[0].Id)
	s.Require().EqualValues(1, balances[0].AddressID)
	s.Require().EqualValues(3, balances[0].ContractID)
	s.Require().True(balances[0].TokenID.Equal(decimal.NewFromInt(0)))
}

// TestTokenBalanceFilterWithLimit tests Filter with limit
func (s *StorageTestSuite) TestTokenBalanceFilterWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  3,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 3)
}

// TestTokenBalanceFilterWithOffset tests Filter with offset
func (s *StorageTestSuite) TestTokenBalanceFilterWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  3,
		Offset: 2,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 3)
}

// TestTokenBalanceFilterWithLimitOffset tests Filter with both limit and offset
func (s *StorageTestSuite) TestTokenBalanceFilterWithLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	allBalances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  15,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(allBalances, 15)

	// Debug: print all IDs and balances
	for i, b := range allBalances {
		s.T().Logf("allBalances[%d]: id=%d, balance=%s", i, b.Id, b.Balance.String())
	}

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  2,
		Offset: 3,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 2)

	s.T().Logf("balances[0]: id=%d, balance=%s", balances[0].Id, balances[0].Balance.String())
	s.T().Logf("balances[1]: id=%d, balance=%s", balances[1].Id, balances[1].Balance.String())

	// Verify pagination consistency
	s.Require().EqualValues(allBalances[3].Id, balances[0].Id)
	s.Require().EqualValues(allBalances[4].Id, balances[1].Id)
}

// TestTokenBalanceFilterSortDesc tests Filter with descending sort
func (s *StorageTestSuite) TestTokenBalanceFilterSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   "desc",
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 5)

	// Check descending order by balance
	for i := 1; i < len(balances); i++ {
		s.Require().True(balances[i-1].Balance.GreaterThanOrEqual(balances[i].Balance))
	}
}

// TestTokenBalanceFilterSortAsc tests Filter with explicit ascending sort
func (s *StorageTestSuite) TestTokenBalanceFilterSortAsc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   "asc",
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 5)

	// Check ascending order by balance
	for i := 1; i < len(balances); i++ {
		s.Require().True(balances[i-1].Balance.LessThanOrEqual(balances[i].Balance))
	}
}

// TestTokenBalanceFilterInvalidSort tests Filter with invalid sort value (should default to asc)
func (s *StorageTestSuite) TestTokenBalanceFilterInvalidSort() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  3,
		Offset: 0,
		Sort:   "invalid",
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 3)

	// Should default to ascending order by balance
	for i := 1; i < len(balances); i++ {
		s.Require().True(balances[i-1].Balance.LessThanOrEqual(balances[i].Balance))
	}
}

// TestTokenBalanceFilterNoResults tests Filter with no matching results
func (s *StorageTestSuite) TestTokenBalanceFilterNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(999)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId: &addressId,
		Limit:     10,
		Offset:    0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 0)
}

// TestTokenBalanceFilterOffsetExceedsTotal tests Filter with offset exceeding total
func (s *StorageTestSuite) TestTokenBalanceFilterOffsetExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  10,
		Offset: 100,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 0)
}

// TestTokenBalanceFilterJoinFields tests that JOIN fields are populated correctly
func (s *StorageTestSuite) TestTokenBalanceFilterJoinFields() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  1,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 1)

	// Check JOIN fields are populated
	s.Require().NotNil(balances[0].Contract.Address)
	s.Require().NotEmpty(balances[0].Contract.Address.Hash)
	s.Require().NotNil(balances[0].Address)
	s.Require().NotEmpty(balances[0].Address.Hash)
}

// TestTokenBalanceFilterZeroLimit tests Filter with zero limit (should use default 10)
func (s *StorageTestSuite) TestTokenBalanceFilterZeroLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  0,
		Offset: 0,
	})
	s.Require().NoError(err)
	// Should use default limit 10
	s.Require().Len(balances, 10)
}

// TestTokenBalanceFilterLimitExceedsMax tests Filter with limit > 100 (should use default 10)
func (s *StorageTestSuite) TestTokenBalanceFilterLimitExceedsMax() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  200,
		Offset: 0,
	})
	s.Require().NoError(err)
	// Should be limited to default 10
	s.Require().Len(balances, 10)
}

// TestTokenBalanceFilterNegativeLimit tests Filter with negative limit (should use default 10)
func (s *StorageTestSuite) TestTokenBalanceFilterNegativeLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		Limit:  -5,
		Offset: 0,
	})
	s.Require().NoError(err)
	// Should use default limit 10
	s.Require().Len(balances, 10)
}

// TestTokenBalanceFilterDifferentTokens tests filtering different token_ids
func (s *StorageTestSuite) TestTokenBalanceFilterDifferentTokens() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		tokenId       int64
		expectedCount int
	}{
		{0, 8},
		{1, 2},
		{2, 2},
		{5, 2},
		{10, 1},
	}

	for _, tc := range testCases {
		tokenId := decimal.NewFromInt(tc.tokenId)
		balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
			TokenId: &tokenId,
			Limit:   15,
			Offset:  0,
		})
		s.Require().NoError(err)
		s.Require().Len(balances, tc.expectedCount, "token_id: %d", tc.tokenId)

		// Verify all have correct token_id
		for _, balance := range balances {
			s.Require().True(balance.TokenID.Equal(decimal.NewFromInt(tc.tokenId)))
		}
	}
}

// TestTokenBalanceFilterMultipleAddresses tests balances across multiple addresses
func (s *StorageTestSuite) TestTokenBalanceFilterMultipleAddresses() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Get balances for address 1
	addressId1 := uint64(1)
	balances1, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId: &addressId1,
		Limit:     10,
		Offset:    0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances1, 5)

	// Get balances for address 2
	addressId2 := uint64(2)
	balances2, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId: &addressId2,
		Limit:     10,
		Offset:    0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances2, 3)

	// Verify no overlap
	for _, b1 := range balances1 {
		for _, b2 := range balances2 {
			s.Require().NotEqual(b1.Id, b2.Id)
		}
	}
}

// TestTokenBalanceFilterMultipleContracts tests balances across multiple contracts
func (s *StorageTestSuite) TestTokenBalanceFilterMultipleContracts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		contractId    uint64
		expectedCount int
	}{
		{3, 10},
		{4, 2},
		{5, 2},
		{6, 1},
	}

	for _, tc := range testCases {
		balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
			ContractId: &tc.contractId,
			Limit:      15,
			Offset:     0,
		})
		s.Require().NoError(err)
		s.Require().Len(balances, tc.expectedCount, "contract_id: %d", tc.contractId)

		// Verify all have correct contract_id
		for _, balance := range balances {
			s.Require().EqualValues(tc.contractId, balance.ContractID)
		}
	}
}

// TestTokenBalanceFilterBalanceValues tests that balance values are correct
func (s *StorageTestSuite) TestTokenBalanceFilterBalanceValues() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	contractId := uint64(3)
	tokenId := decimal.NewFromInt(0)
	balances, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		AddressId:  &addressId,
		ContractId: &contractId,
		TokenId:    &tokenId,
		Limit:      10,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 1)

	// Check specific balance value
	s.Require().True(balances[0].Balance.Equal(decimal.NewFromInt(1000000000000000000)))
}

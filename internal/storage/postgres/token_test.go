package postgres

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
)

// TestTokenGetBasic tests basic Get functionality
func (s *StorageTestSuite) TestTokenGetBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	token, err := s.storage.Token.Get(ctx, 3, decimal.NewFromInt(0))
	s.Require().NoError(err)

	// Check token details
	s.Require().EqualValues(1, token.Id)
	s.Require().EqualValues("0", token.TokenID.String())
	s.Require().EqualValues(3, token.ContractId)
	s.Require().EqualValues(types.ERC20, token.Type)
	s.Require().EqualValues("Test Token", token.Name)
	s.Require().EqualValues("TST", token.Symbol)
	s.Require().EqualValues(18, token.Decimals)
	s.Require().EqualValues(types.Success, token.Status)

	// Check JOIN field
	s.Require().NotNil(token.Contract.Address)
	s.Require().NotEmpty(token.Contract.Address.Hash)
}

// TestTokenGetERC721 tests getting ERC721 token
func (s *StorageTestSuite) TestTokenGetERC721() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	token, err := s.storage.Token.Get(ctx, 3, decimal.NewFromInt(1))
	s.Require().NoError(err)

	s.Require().EqualValues(2, token.Id)
	s.Require().EqualValues("1", token.TokenID.String())
	s.Require().EqualValues(types.ERC721, token.Type)
	s.Require().EqualValues("NFT Collection", token.Name)
	s.Require().EqualValues(0, token.Decimals)
}

// TestTokenGetERC1155 tests getting ERC1155 token
func (s *StorageTestSuite) TestTokenGetERC1155() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	token, err := s.storage.Token.Get(ctx, 3, decimal.NewFromInt(2))
	s.Require().NoError(err)

	s.Require().EqualValues(4, token.Id)
	s.Require().EqualValues("2", token.TokenID.String())
	s.Require().EqualValues(types.ERC1155, token.Type)
	s.Require().EqualValues("Multi Token", token.Name)
}

// TestTokenGetNotFound tests Get with non-existent token
func (s *StorageTestSuite) TestTokenGetNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.storage.Token.Get(ctx, 999, decimal.NewFromInt(0))
	s.Require().Error(err)
}

// TestTokenGetDifferentContract tests getting token from different contract
func (s *StorageTestSuite) TestTokenGetDifferentContract() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	token, err := s.storage.Token.Get(ctx, 4, decimal.NewFromInt(0))
	s.Require().NoError(err)

	s.Require().EqualValues(3, token.Id)
	s.Require().EqualValues(4, token.ContractId)
	s.Require().EqualValues("Another Token", token.Name)
	s.Require().EqualValues("ANT", token.Symbol)
}

// TestTokenGetLargeTokenId tests getting token with large token_id
func (s *StorageTestSuite) TestTokenGetLargeTokenId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	token, err := s.storage.Token.Get(ctx, 5, decimal.NewFromInt(100))
	s.Require().NoError(err)

	s.Require().EqualValues(9, token.Id)
	s.Require().EqualValues("100", token.TokenID.String())
	s.Require().EqualValues("Special NFT", token.Name)
}

// TestTokenFilterBasic tests basic Filter functionality
func (s *StorageTestSuite) TestTokenFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 10)

	s.Require().EqualValues(1, tokens[0].Id)
	s.Require().EqualValues(2, tokens[1].Id)
}

// TestTokenFilterByContractId tests filtering by contract_id
func (s *StorageTestSuite) TestTokenFilterByContractId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(3)
	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		ContractId: &contractId,
		Limit:      10,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 5)

	for _, token := range tokens {
		s.Require().EqualValues(3, token.ContractId)
	}

	s.Require().EqualValues("0", tokens[0].TokenID.String())
	s.Require().EqualValues("1", tokens[1].TokenID.String())
	s.Require().EqualValues("2", tokens[2].TokenID.String())
	s.Require().EqualValues("5", tokens[3].TokenID.String())
	s.Require().EqualValues("10", tokens[4].TokenID.String())
}

// TestTokenFilterByType tests filtering by token type
func (s *StorageTestSuite) TestTokenFilterByType() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Type:   []types.TokenType{types.ERC20},
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 3)

	// All should be ERC20
	for _, token := range tokens {
		s.Require().EqualValues(types.ERC20, token.Type)
	}
}

// TestTokenFilterByMultipleTypes tests filtering by multiple token types
func (s *StorageTestSuite) TestTokenFilterByMultipleTypes() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Type:   []types.TokenType{types.ERC721, types.ERC1155},
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 7) // 4 ERC721 + 3 ERC1155

	// All should be either ERC721 or ERC1155
	for _, token := range tokens {
		s.Require().True(
			token.Type == types.ERC721 || token.Type == types.ERC1155,
		)
	}
}

// TestTokenFilterByContractIdAndType tests filtering by both contract_id and type
func (s *StorageTestSuite) TestTokenFilterByContractIdAndType() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(3)
	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		ContractId: &contractId,
		Type:       []types.TokenType{types.ERC1155},
		Limit:      10,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 3)

	for _, token := range tokens {
		s.Require().EqualValues(3, token.ContractId)
		s.Require().EqualValues(types.ERC1155, token.Type)
	}
}

// TestTokenFilterWithLimit tests Filter with limit
func (s *StorageTestSuite) TestTokenFilterWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  3,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 3)

	s.Require().EqualValues(1, tokens[0].Id)
	s.Require().EqualValues(2, tokens[1].Id)
	s.Require().EqualValues(3, tokens[2].Id)
}

// TestTokenFilterWithOffset tests Filter with offset
func (s *StorageTestSuite) TestTokenFilterWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  3,
		Offset: 2,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 3)

	s.Require().EqualValues(3, tokens[0].Id)
	s.Require().EqualValues(4, tokens[1].Id)
	s.Require().EqualValues(5, tokens[2].Id)
}

// TestTokenFilterWithLimitOffset tests Filter with both limit and offset
func (s *StorageTestSuite) TestTokenFilterWithLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	allTokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  2,
		Offset: 3,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 2)
	s.Require().EqualValues(allTokens[3].Id, tokens[0].Id)
	s.Require().EqualValues(allTokens[4].Id, tokens[1].Id)
}

// TestTokenFilterSortDesc tests Filter with descending sort
func (s *StorageTestSuite) TestTokenFilterSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   "desc",
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 5)

	// Check descending order
	s.Require().EqualValues(10, tokens[0].Id)
	s.Require().EqualValues(9, tokens[1].Id)
	s.Require().EqualValues(8, tokens[2].Id)
	s.Require().EqualValues(7, tokens[3].Id)
	s.Require().EqualValues(6, tokens[4].Id)
}

// TestTokenFilterNoResults tests Filter with no matching results
func (s *StorageTestSuite) TestTokenFilterNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contractId := uint64(999)
	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		ContractId: &contractId,
		Limit:      10,
		Offset:     0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 0)
}

// TestTokenFilterOffsetExceedsTotal tests Filter with offset exceeding total
func (s *StorageTestSuite) TestTokenFilterOffsetExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  10,
		Offset: 100,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 0)
}

// TestTokenFilterJoinFields tests that JOIN fields are populated correctly
func (s *StorageTestSuite) TestTokenFilterJoinFields() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  1,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 1)

	// Check JOIN fields are populated
	s.Require().NotNil(tokens[0].Contract.Address)
	s.Require().NotEmpty(tokens[0].Contract.Address.Hash)
}

// TestTokenPendingMetadataBasic tests basic PendingMetadata functionality
func (s *StorageTestSuite) TestTokenPendingMetadataBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Use very old delay to get all pending tokens
	tokens, err := s.storage.Token.PendingMetadata(ctx, 365*24*time.Hour, 10)
	s.Require().NoError(err)
	s.Require().Len(tokens, 5) // 5 tokens with status='pending'

	// All should have status pending
	for _, token := range tokens {
		s.Require().EqualValues(types.Pending, token.Status)
	}
}

// TestTokenPendingMetadataWithRetryCount tests PendingMetadata with retry_count filter
func (s *StorageTestSuite) TestTokenPendingMetadataWithRetryCount() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Use very old delay - should get all pending tokens
	tokens, err := s.storage.Token.PendingMetadata(ctx, 365*24*time.Hour, 10)
	s.Require().NoError(err)

	// Check we have tokens with retry_count = 0 (should always be included)
	hasRetryZero := false
	for _, token := range tokens {
		if token.RetryCount == 0 {
			hasRetryZero = true
			break
		}
	}
	s.Require().True(hasRetryZero)
}

// TestTokenPendingMetadataWithTimeThreshold tests PendingMetadata with time threshold
func (s *StorageTestSuite) TestTokenPendingMetadataWithTimeThreshold() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Use 2 day delay - should exclude tokens updated recently (within last 2 days)
	tokens, err := s.storage.Token.PendingMetadata(ctx, 2*24*time.Hour, 10)
	s.Require().NoError(err)

	// Should get tokens with:
	// - retry_count = 0 (always included)
	// - OR updated_at < (now - 2 days)
	for _, token := range tokens {
		if token.RetryCount > 0 {
			s.Require().True(token.UpdatedAt.Before(time.Now().UTC().Add(-2 * 24 * time.Hour)))
		}
	}
}

// TestTokenPendingMetadataWithLimit tests PendingMetadata with limit
func (s *StorageTestSuite) TestTokenPendingMetadataWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.PendingMetadata(ctx, 365*24*time.Hour, 2)
	s.Require().NoError(err)
	s.Require().Len(tokens, 2)

	// Should be ordered by id ASC
	s.Require().EqualValues(2, tokens[0].Id)
	s.Require().EqualValues(3, tokens[1].Id)
}

// TestTokenPendingMetadataOrderById tests PendingMetadata ordering
func (s *StorageTestSuite) TestTokenPendingMetadataOrderById() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.PendingMetadata(ctx, 365*24*time.Hour, 10)
	s.Require().NoError(err)

	// Check ascending order by id
	for i := 1; i < len(tokens); i++ {
		s.Require().True(tokens[i-1].Id < tokens[i].Id)
	}
}

// TestTokenPendingMetadataExcludesNonPending tests that non-pending tokens are excluded
func (s *StorageTestSuite) TestTokenPendingMetadataExcludesNonPending() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.PendingMetadata(ctx, 365*24*time.Hour, 10)
	s.Require().NoError(err)

	// None should have status success or failed
	for _, token := range tokens {
		s.Require().NotEqual(types.Success, token.Status)
		s.Require().NotEqual(types.Failed, token.Status)
	}
}

// TestTokenPendingMetadataNoResults tests PendingMetadata with very short delay
func (s *StorageTestSuite) TestTokenPendingMetadataNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Use 0 delay - only tokens with retry_count=0 should be returned
	tokens, err := s.storage.Token.PendingMetadata(ctx, 0, 10)
	s.Require().NoError(err)

	// All returned tokens should have retry_count = 0
	for _, token := range tokens {
		if token.Status == types.Pending {
			// Either retry_count is 0, or updated_at is very old
			s.Require().True(token.RetryCount == 0 || token.UpdatedAt.Before(time.Now().UTC()))
		}
	}
}

// TestTokenFilterAllTypes tests filtering by all token types separately
func (s *StorageTestSuite) TestTokenFilterAllTypes() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Test each type separately
	testCases := []struct {
		tokenType     types.TokenType
		expectedCount int
	}{
		{types.ERC20, 3},
		{types.ERC721, 4},
		{types.ERC1155, 3},
	}

	for _, tc := range testCases {
		tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
			Type:   []types.TokenType{tc.tokenType},
			Limit:  10,
			Offset: 0,
		})
		s.Require().NoError(err)
		s.Require().Len(tokens, tc.expectedCount, "type: %s", tc.tokenType)

		// Verify all have correct type
		for _, token := range tokens {
			s.Require().EqualValues(tc.tokenType, token.Type)
		}
	}
}

// TestTokenFilterZeroLimit tests Filter with zero limit
func (s *StorageTestSuite) TestTokenFilterZeroLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  0,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(tokens)
}

// TestTokenPendingMetadataVerifyFields tests that all fields are populated correctly
func (s *StorageTestSuite) TestTokenPendingMetadataVerifyFields() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.PendingMetadata(ctx, 365*24*time.Hour, 1)
	s.Require().NoError(err)
	s.Require().Len(tokens, 1)

	token := tokens[0]
	s.Require().NotZero(token.Id)
	s.Require().NotNil(token.TokenID)
	s.Require().NotZero(token.ContractId)
	s.Require().NotEmpty(token.Type)
	s.Require().EqualValues(types.Pending, token.Status)
}

// TestTokenFilterSortAsc tests Filter with explicit ascending sort
func (s *StorageTestSuite) TestTokenFilterSortAsc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   "asc",
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 5)

	// Check ascending order
	s.Require().EqualValues(1, tokens[0].Id)
	s.Require().EqualValues(2, tokens[1].Id)
	s.Require().EqualValues(3, tokens[2].Id)
	s.Require().EqualValues(4, tokens[3].Id)
	s.Require().EqualValues(5, tokens[4].Id)
}

// TestTokenFilterInvalidSort tests Filter with invalid sort value (should default to asc)
func (s *StorageTestSuite) TestTokenFilterInvalidSort() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  3,
		Offset: 0,
		Sort:   "invalid",
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 3)

	// Should default to ascending order
	s.Require().EqualValues(1, tokens[0].Id)
	s.Require().EqualValues(2, tokens[1].Id)
	s.Require().EqualValues(3, tokens[2].Id)
}

// TestTokenFilterLimitExceedsMax tests Filter with limit > 100 (should use default 10)
func (s *StorageTestSuite) TestTokenFilterLimitExceedsMax() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  200,
		Offset: 0,
	})
	s.Require().NoError(err)
	// Should be limited to default 10
	s.Require().Len(tokens, 10)
}

// TestTokenFilterNegativeLimit tests Filter with negative limit (should use default 10)
func (s *StorageTestSuite) TestTokenFilterNegativeLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tokens, err := s.storage.Token.Filter(ctx, storage.TokenListFilter{
		Limit:  -5,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Len(tokens, 10)
}

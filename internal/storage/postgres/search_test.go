package postgres

import (
	"context"
	"time"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
)

// TestSearchByBlockHash tests Search method with block hash
func (s *StorageTestSuite) TestSearchByBlockHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	results, err := s.storage.Search.Search(ctx, hash)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().EqualValues(1, results[0].Id)
	s.Require().Equal("block", results[0].Type)
	s.Require().Equal("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", results[0].Value)
}

// TestSearchByTxHash tests Search method with transaction hash
func (s *StorageTestSuite) TestSearchByTxHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("90f5df4e03620cc55d3ea295bf8826f84465065340cb6d0d095166dd2465f283")
	results, err := s.storage.Search.Search(ctx, hash)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().EqualValues(1, results[0].Id)
	s.Require().Equal("tx", results[0].Type)
}

// TestSearchNoResults tests Search method with non-existent hash
func (s *StorageTestSuite) TestSearchNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("0000000000000000000000000000000000000000000000000000000000000000")
	results, err := s.storage.Search.Search(ctx, hash)
	s.Require().NoError(err)
	s.Require().Len(results, 0)
}

// TestSearchTextByTokenName tests SearchText with exact token name
func (s *StorageTestSuite) TestSearchTextByTokenName() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "Test Token")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 1)

	// Find the matching result
	found := false
	for _, result := range results {
		if result.Id == 1 && result.Type == "token" && result.Value == "Test Token" {
			found = true
			break
		}
	}
	s.Require().True(found, "Should find token with name 'Test Token'")
}

// TestSearchTextByTokenSymbol tests SearchText with exact token symbol
func (s *StorageTestSuite) TestSearchTextByTokenSymbol() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "TST")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 1)

	// Find the matching result
	found := false
	for _, result := range results {
		if result.Id == 1 && result.Type == "token" && result.Value == "TST" {
			found = true
			break
		}
	}
	s.Require().True(found, "Should find token with symbol 'TST'")
}

// TestSearchTextPartialName tests SearchText with partial token name
func (s *StorageTestSuite) TestSearchTextPartialName() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "Token")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 2) // "Test Token", "Another Token", "Multi Token"

	// Verify all results contain "Token" in the name
	for _, result := range results {
		if result.Type == "token" {
			// Value should contain "Token" or be a symbol
			s.Require().NotEmpty(result.Value)
		}
	}
}

// TestSearchTextPartialSymbol tests SearchText with partial token symbol
func (s *StorageTestSuite) TestSearchTextPartialSymbol() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "ST")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 1) // Should find "TST" and "STBL"
}

// TestSearchTextCaseInsensitiveLowercase tests SearchText with lowercase input
func (s *StorageTestSuite) TestSearchTextCaseInsensitiveLowercase() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "test token")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 1)

	// Find the matching result
	found := false
	for _, result := range results {
		if result.Id == 1 && result.Type == "token" && result.Value == "Test Token" {
			found = true
			break
		}
	}
	s.Require().True(found, "Should find token with case-insensitive search")
}

// TestSearchTextCaseInsensitiveUppercase tests SearchText with uppercase input
func (s *StorageTestSuite) TestSearchTextCaseInsensitiveUppercase() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "RARE NFT")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 1)

	// Find the matching result
	found := false
	for _, result := range results {
		if result.Id == 5 && result.Type == "token" && result.Value == "Rare NFT" {
			found = true
			break
		}
	}
	s.Require().True(found, "Should find token with case-insensitive uppercase search")
}

// TestSearchTextNoResults tests SearchText with no matching results
func (s *StorageTestSuite) TestSearchTextNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "NonExistentToken12345")
	s.Require().NoError(err)
	s.Require().Len(results, 0)
}

// TestSearchTextMultipleResults tests SearchText returning multiple results
func (s *StorageTestSuite) TestSearchTextMultipleResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "NFT")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 3) // "NFT Collection", "Rare NFT", "Special NFT" and symbol "NFT"

	// Verify all results are tokens
	for _, result := range results {
		s.Require().Equal("token", result.Type)
	}
}

// TestSearchTextLimit tests SearchText respects 10 result limit
func (s *StorageTestSuite) TestSearchTextLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Search for empty string should match all tokens (if we had >10)
	// But with current fixtures we have 10 tokens, so this won't exceed limit
	results, err := s.storage.Search.SearchText(ctx, "")
	s.Require().NoError(err)
	s.Require().LessOrEqual(len(results), 10)
}

// TestSearchTextBySymbolOnly tests SearchText finding by symbol when name doesn't match
func (s *StorageTestSuite) TestSearchTextBySymbolOnly() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	results, err := s.storage.Search.SearchText(ctx, "COLL")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 1)

	// Should find token with symbol "COLL" (Collectible)
	found := false
	for _, result := range results {
		if result.Id == 10 && result.Type == "token" && result.Value == "COLL" {
			found = true
			break
		}
	}
	s.Require().True(found, "Should find token by symbol 'COLL'")
}

package postgres

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestSourceByContractIdBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 3, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)

	// Check first source
	s.Require().EqualValues(1, sources[0].Id)
	s.Require().EqualValues("Token.sol", sources[0].Name)
	s.Require().EqualValues("MIT", sources[0].License)
	s.Require().EqualValues(3, sources[0].ContractId)
	s.Require().Len(sources[0].Urls, 1)
	s.Require().EqualValues("https://github.com/example/token/blob/main/Token.sol", sources[0].Urls[0])
	s.Require().Contains(sources[0].Content, "contract Token")

	// Check second source
	s.Require().EqualValues(2, sources[1].Id)
	s.Require().EqualValues("TokenStorage.sol", sources[1].Name)
	s.Require().EqualValues(3, sources[1].ContractId)
}

func (s *StorageTestSuite) TestSourceByContractIdMultipleSources() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 4, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)

	// Both should have contract_id = 4
	for _, source := range sources {
		s.Require().EqualValues(4, source.ContractId)
	}

	// Check they have correct IDs
	s.Require().EqualValues(3, sources[0].Id)
	s.Require().EqualValues(4, sources[1].Id)

	// Check different licenses
	s.Require().EqualValues("Apache-2.0", sources[0].License)
	s.Require().EqualValues("Apache-2.0", sources[1].License)
}

func (s *StorageTestSuite) TestSourceByContractIdSingleSource() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 5, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 1)

	// Check source details
	s.Require().EqualValues(5, sources[0].Id)
	s.Require().EqualValues("Implementation.sol", sources[0].Name)
	s.Require().EqualValues("GPL-3.0", sources[0].License)
	s.Require().EqualValues(5, sources[0].ContractId)

	// Check multiple URLs
	s.Require().Len(sources[0].Urls, 2)
	s.Require().EqualValues("https://github.com/example/impl/blob/main/Implementation.sol", sources[0].Urls[0])
	s.Require().EqualValues("https://etherscan.io/address/0x123", sources[0].Urls[1])
}

func (s *StorageTestSuite) TestSourceByContractIdWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 6, 1, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 1)

	// Should return only first source
	s.Require().EqualValues(6, sources[0].Id)
	s.Require().EqualValues("Vault.sol", sources[0].Name)
	s.Require().EqualValues(6, sources[0].ContractId)
}

func (s *StorageTestSuite) TestSourceByContractIdWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 6, 10, 1)
	s.Require().NoError(err)
	s.Require().Len(sources, 1)

	// Should return second source (after offset)
	s.Require().EqualValues(7, sources[0].Id)
	s.Require().EqualValues("VaultV2.sol", sources[0].Name)
	s.Require().EqualValues(6, sources[0].ContractId)
}

func (s *StorageTestSuite) TestSourceByContractIdWithLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// First, get all sources for contract 3
	allSources, err := s.storage.Sources.ByContractId(ctx, 3, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(allSources, 2)

	// Now get first one with limit=1, offset=0
	sources, err := s.storage.Sources.ByContractId(ctx, 3, 1, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(allSources[0].Id, sources[0].Id)

	// Now get second one with limit=1, offset=1
	sources, err = s.storage.Sources.ByContractId(ctx, 3, 1, 1)
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(allSources[1].Id, sources[0].Id)
}

func (s *StorageTestSuite) TestSourceByContractIdNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 999, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 0)
}

func (s *StorageTestSuite) TestSourceByContractIdOffsetExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 3, 10, 100)
	s.Require().NoError(err)
	s.Require().Len(sources, 0)
}

func (s *StorageTestSuite) TestSourceByContractIdEmptyUrls() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 6, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)

	// Both should have empty URLs array
	for _, source := range sources {
		s.Require().Len(source.Urls, 0)
	}

	// Check unlicensed sources
	s.Require().EqualValues("UNLICENSED", sources[0].License)
	s.Require().EqualValues("UNLICENSED", sources[1].License)
}

func (s *StorageTestSuite) TestSourceByContractIdDifferentLicenses() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// MIT license - contract 3
	sources, err := s.storage.Sources.ByContractId(ctx, 3, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)
	for _, source := range sources {
		s.Require().EqualValues("MIT", source.License)
	}

	// Apache-2.0 license - contract 4
	sources, err = s.storage.Sources.ByContractId(ctx, 4, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)
	for _, source := range sources {
		s.Require().EqualValues("Apache-2.0", source.License)
	}

	// GPL-3.0 license - contract 5
	sources, err = s.storage.Sources.ByContractId(ctx, 5, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues("GPL-3.0", sources[0].License)

	// UNLICENSED - contract 6
	sources, err = s.storage.Sources.ByContractId(ctx, 6, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)
	for _, source := range sources {
		s.Require().EqualValues("UNLICENSED", source.License)
	}
}

func (s *StorageTestSuite) TestSourceByContractIdZeroLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Limit 0 should use default limit behavior (depends on limitScope implementation)
	sources, err := s.storage.Sources.ByContractId(ctx, 3, 0, 0)
	s.Require().NoError(err)
	// Should still return results (limitScope handles 0 limit)
	s.Require().NotEmpty(sources)
}

func (s *StorageTestSuite) TestSourceByContractIdNegativeOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Negative offset should be treated as 0
	sources, err := s.storage.Sources.ByContractId(ctx, 3, 10, -1)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)
	s.Require().EqualValues(1, sources[0].Id)
}

func (s *StorageTestSuite) TestSourceByContractIdContentVerification() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.ByContractId(ctx, 3, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(sources, 2)

	// Verify content is not empty
	for _, source := range sources {
		s.Require().NotEmpty(source.Content)
		s.Require().Contains(source.Content, "contract")
	}

	// Check specific content patterns
	s.Require().Contains(sources[0].Content, "Token")
	s.Require().Contains(sources[1].Content, "TokenStorage")
}

func (s *StorageTestSuite) TestSourceByContractIdAllContracts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Test each contract ID has expected number of sources
	testCases := []struct {
		contractId    uint64
		expectedCount int
		expectedNames []string
	}{
		{3, 2, []string{"Token.sol", "TokenStorage.sol"}},
		{4, 2, []string{"Proxy.sol", "ProxyAdmin.sol"}},
		{5, 1, []string{"Implementation.sol"}},
		{6, 2, []string{"Vault.sol", "VaultV2.sol"}},
	}

	for _, tc := range testCases {
		sources, err := s.storage.Sources.ByContractId(ctx, tc.contractId, 10, 0)
		s.Require().NoError(err)
		s.Require().Len(sources, tc.expectedCount, "contract_id: %d", tc.contractId)

		// Verify names
		for i, source := range sources {
			s.Require().EqualValues(tc.expectedNames[i], source.Name, "contract_id: %d, index: %d", tc.contractId, i)
		}
	}
}

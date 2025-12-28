package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestProxyContractNotResolved() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.NotResolved(ctx)
	s.Require().NoError(err)
	s.Require().Len(proxies, 2)

	// Should return only status = 'new' ordered by height DESC
	s.Require().EqualValues(10, proxies[0].Id)
	s.Require().EqualValues(types.New, proxies[0].Status)
	s.Require().EqualValues(7, proxies[1].Id)
	s.Require().EqualValues(types.New, proxies[1].Status)

	// Check that Contract.Address is loaded
	s.Require().NotNil(proxies[0].Contract.Address.Hash)
}

func (s *StorageTestSuite) TestProxyContractFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 7)

	// Check first proxy (sorted by height ASC)
	s.Require().EqualValues(6, proxies[0].Id)
	s.Require().EqualValues(400, proxies[0].Height)
	s.Require().EqualValues(types.EIP1967, proxies[0].Type)
	s.Require().EqualValues(types.Resolved, proxies[0].Status)

	// Verify JOIN fields are loaded
	s.Require().NotNil(proxies[0].Implementation.Address.Hash)
	s.Require().NotNil(proxies[0].Contract.Address.Hash)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", proxies[0].Implementation.Address.Hash.Hex())
	s.Require().EqualValues("0x60f055506ba543ea0942dc8ca03f596ab75bc882", proxies[0].Contract.Address.Hash.Hex())
}

func (s *StorageTestSuite) TestProxyContractFilterDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 7)

	// Check descending order by height
	s.Require().EqualValues(12, proxies[0].Id)
	s.Require().EqualValues(800, proxies[0].Height)
	s.Require().EqualValues(6, proxies[6].Id)
	s.Require().EqualValues(400, proxies[6].Height)
}

func (s *StorageTestSuite) TestProxyContractFilterByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: 600,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 2)

	// Both should be at height 600
	for _, proxy := range proxies {
		s.Require().EqualValues(600, proxy.Height)
	}

	s.Require().EqualValues(9, proxies[0].Id)
	s.Require().EqualValues(10, proxies[1].Id)
}

func (s *StorageTestSuite) TestProxyContractFilterByImplementationId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:            10,
		Offset:           0,
		Sort:             sdk.SortOrderAsc,
		ImplementationId: 3,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 2)

	// Both should have implementation_id = 3
	for _, proxy := range proxies {
		s.Require().NotNil(proxy.ImplementationID)
		s.Require().EqualValues(3, *proxy.ImplementationID)
	}

	s.Require().EqualValues(6, proxies[0].Id)
	s.Require().EqualValues(12, proxies[1].Id)

	// Verify implementation address
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", proxies[0].Implementation.Address.Hash.Hex())
}

func (s *StorageTestSuite) TestProxyContractFilterByTypeSingle() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Type:   []types.ProxyType{types.EIP1967},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 3)

	// All should be EIP1967
	for _, proxy := range proxies {
		s.Require().EqualValues(types.EIP1967, proxy.Type)
	}

	s.Require().EqualValues(6, proxies[0].Id)
	s.Require().EqualValues(8, proxies[1].Id)
	s.Require().EqualValues(11, proxies[2].Id)
}

func (s *StorageTestSuite) TestProxyContractFilterByTypeMultiple() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Type:   []types.ProxyType{types.EIP1967, types.EIP1167},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 4)

	// Should contain both EIP1967 and EIP1167
	s.Require().EqualValues(types.EIP1967, proxies[0].Type)
	s.Require().EqualValues(types.EIP1167, proxies[1].Type)
	s.Require().EqualValues(types.EIP1967, proxies[2].Type)
	s.Require().EqualValues(types.EIP1967, proxies[3].Type)
}

func (s *StorageTestSuite) TestProxyContractFilterByStatusSingle() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Status: []types.ProxyStatus{types.Resolved},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 4)

	// All should be resolved
	for _, proxy := range proxies {
		s.Require().EqualValues(types.Resolved, proxy.Status)
	}

	s.Require().EqualValues(6, proxies[0].Id)
	s.Require().EqualValues(8, proxies[1].Id)
	s.Require().EqualValues(11, proxies[2].Id)
	s.Require().EqualValues(12, proxies[3].Id)
}

func (s *StorageTestSuite) TestProxyContractFilterByStatusMultiple() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Status: []types.ProxyStatus{types.New, types.Error},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 3)

	// Should contain new and error statuses
	s.Require().EqualValues(types.New, proxies[0].Status)
	s.Require().EqualValues(types.Error, proxies[1].Status)
	s.Require().EqualValues(types.New, proxies[2].Status)
}

func (s *StorageTestSuite) TestProxyContractFilterCombined() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: 500,
		Type:   []types.ProxyType{types.EIP1967},
		Status: []types.ProxyStatus{types.Resolved},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 1)

	// Should match all filters
	s.Require().EqualValues(8, proxies[0].Id)
	s.Require().EqualValues(500, proxies[0].Height)
	s.Require().EqualValues(types.EIP1967, proxies[0].Type)
	s.Require().EqualValues(types.Resolved, proxies[0].Status)
}

func (s *StorageTestSuite) TestProxyContractFilterLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  2,
		Offset: 2,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 2)

	// Should skip first two and return third and fourth
	s.Require().EqualValues(8, proxies[0].Id)
	s.Require().EqualValues(9, proxies[1].Id)
}

func (s *StorageTestSuite) TestProxyContractFilterMinimalLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  1,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 1)

	// Should return only first
	s.Require().EqualValues(6, proxies[0].Id)

	// Verify JOIN fields work with limit=1
	s.Require().NotNil(proxies[0].Implementation.Address.Hash)
	s.Require().NotNil(proxies[0].Contract.Address.Hash)
}

func (s *StorageTestSuite) TestProxyContractFilterExactBoundary() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 6,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 1)

	// Should return only last proxy
	s.Require().EqualValues(12, proxies[0].Id)
}

func (s *StorageTestSuite) TestProxyContractFilterEmptyResult() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 100,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 0)
}

func (s *StorageTestSuite) TestProxyContractFilterNonExistentHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: 999,
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 0)
}

func (s *StorageTestSuite) TestProxyContractFilterDescWithType() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
		Type:   []types.ProxyType{types.EIP1967},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 3)

	// Check descending order
	s.Require().EqualValues(11, proxies[0].Id)
	s.Require().EqualValues(700, proxies[0].Height)
	s.Require().EqualValues(8, proxies[1].Id)
	s.Require().EqualValues(6, proxies[2].Id)

	// Verify JOIN fields
	s.Require().NotNil(proxies[0].Implementation.Address.Hash)
	s.Require().NotNil(proxies[0].Contract.Address.Hash)
}

func (s *StorageTestSuite) TestProxyContractFilterAllFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:            10,
		Offset:           0,
		Sort:             sdk.SortOrderAsc,
		Height:           400,
		ImplementationId: 3,
		Type:             []types.ProxyType{types.EIP1967},
		Status:           []types.ProxyStatus{types.Resolved},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 1)

	// Verify all filters are applied
	proxy := proxies[0]
	s.Require().EqualValues(6, proxy.Id)
	s.Require().EqualValues(400, proxy.Height)
	s.Require().NotNil(proxy.ImplementationID)
	s.Require().EqualValues(3, *proxy.ImplementationID)
	s.Require().EqualValues(types.EIP1967, proxy.Type)
	s.Require().EqualValues(types.Resolved, proxy.Status)
}

func (s *StorageTestSuite) TestProxyContractFilterHeightAndStatus() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: 600,
		Status: []types.ProxyStatus{types.New},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 1)

	// Should match both filters
	s.Require().EqualValues(10, proxies[0].Id)
	s.Require().EqualValues(600, proxies[0].Height)
	s.Require().EqualValues(types.New, proxies[0].Status)
}

func (s *StorageTestSuite) TestProxyContractFilterTypeAndImplementation() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:            10,
		Offset:           0,
		Sort:             sdk.SortOrderAsc,
		ImplementationId: 3,
		Type:             []types.ProxyType{types.EIP1967, types.CloneWithImmutableArgs},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 2)

	// Both should have implementation_id = 3 and matching types
	s.Require().EqualValues(6, proxies[0].Id)
	s.Require().EqualValues(types.EIP1967, proxies[0].Type)
	s.Require().EqualValues(12, proxies[1].Id)
	s.Require().EqualValues(types.CloneWithImmutableArgs, proxies[1].Type)
}

func (s *StorageTestSuite) TestProxyContractFilterEmptySort() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   "", // Empty sort - should not add ORDER BY to outer query
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 7)

	// Even with empty sort, should return all proxies
	// Order might not be guaranteed, so just check we got them all
	s.Require().NotEmpty(proxies)
}

func (s *StorageTestSuite) TestProxyContractFilterNullImplementation() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Status: []types.ProxyStatus{types.New, types.Error},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 3)

	// Check proxies with NULL implementation_id
	for _, proxy := range proxies {
		if proxy.Id == 7 || proxy.Id == 9 || proxy.Id == 10 {
			s.Require().Nil(proxy.ImplementationID)
			// Implementation should be nil or empty when implementation_id is NULL
			if proxy.Implementation != nil {
				s.Require().EqualValues(0, proxy.Implementation.Id)
			}
		}
	}

	// Verify Contract.Address is still loaded even with NULL implementation
	s.Require().NotNil(proxies[0].Contract.Address.Hash)
}

func (s *StorageTestSuite) TestProxyContractFilterSortingConsistency() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// This test ensures sorting is consistent across multiple calls
	// and verifies the sorting bug fix
	proxies1, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)

	proxies2, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)

	// Both queries should return same order
	s.Require().Len(proxies1, len(proxies2))
	for i := range proxies1 {
		s.Require().EqualValues(proxies1[i].Id, proxies2[i].Id)
		s.Require().EqualValues(proxies1[i].Height, proxies2[i].Height)
	}

	// Verify ascending order by height
	for i := 1; i < len(proxies1); i++ {
		s.Require().GreaterOrEqual(proxies1[i].Height, proxies1[i-1].Height)
	}
}

func (s *StorageTestSuite) TestProxyContractFilterWithJoinsVerification() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Test that JOIN fields are correctly loaded for various cases
	proxies, err := s.storage.ProxyContracts.FilteredList(ctx, storage.ListProxyFilters{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Status: []types.ProxyStatus{types.Resolved},
	})
	s.Require().NoError(err)
	s.Require().Len(proxies, 4)

	// All resolved proxies should have implementation
	for _, proxy := range proxies {
		s.Require().NotNil(proxy.ImplementationID)
		s.Require().NotNil(proxy.Implementation)
		s.Require().NotNil(proxy.Implementation.Address)
		s.Require().NotNil(proxy.Implementation.Address.Hash)
		s.Require().Greater(len(proxy.Implementation.Address.Hash), 0)

		// Contract address should always be loaded
		s.Require().NotNil(proxy.Contract.Address)
		s.Require().NotNil(proxy.Contract.Address.Hash)
		s.Require().Greater(len(proxy.Contract.Address.Hash), 0)
	}
}

package postgres

import (
	"context"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestSourceFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Limit:      10,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 2)

	s.Require().EqualValues(1, sources[0].Id)
	s.Require().EqualValues("Token.sol", sources[0].Name)
	s.Require().EqualValues("MIT", sources[0].License)
	s.Require().EqualValues(3, sources[0].ContractId)
	s.Require().Contains(sources[0].Content, "contract Token")

	s.Require().EqualValues(2, sources[1].Id)
	s.Require().EqualValues("TokenStorage.sol", sources[1].Name)
	s.Require().EqualValues(3, sources[1].ContractId)
}

func (s *StorageTestSuite) TestSourceFilterSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Limit:      10,
		Sort:       sdk.SortOrderDesc,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 2)

	s.Require().EqualValues(2, sources[0].Id)
	s.Require().EqualValues(1, sources[1].Id)
}

func (s *StorageTestSuite) TestSourceFilterWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Limit:      1,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(1, sources[0].Id)
}

func (s *StorageTestSuite) TestSourceFilterWithOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Limit:      10,
		Offset:     1,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(2, sources[0].Id)
}

func (s *StorageTestSuite) TestSourceFilterWithCursorAsc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// CursorID=1 in ASC order should return items with id > 1
	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Limit:      10,
		Sort:       sdk.SortOrderAsc,
		CursorID:   1,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(2, sources[0].Id)
}

func (s *StorageTestSuite) TestSourceFilterWithCursorDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// CursorID=2 in DESC order should return items with id < 2
	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Limit:      10,
		Sort:       sdk.SortOrderDesc,
		CursorID:   2,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(1, sources[0].Id)
}

func (s *StorageTestSuite) TestSourceFilterCursorSkipsOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// When cursor is set, offset should be ignored
	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Limit:      10,
		Offset:     100, // should be ignored
		Sort:       sdk.SortOrderAsc,
		CursorID:   1,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(2, sources[0].Id)
}

func (s *StorageTestSuite) TestSourceFilterNoResults() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 999,
		Limit:      10,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 0)
}

func (s *StorageTestSuite) TestSourceFilterSingleSource() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 5,
		Limit:      10,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 1)
	s.Require().EqualValues(5, sources[0].Id)
	s.Require().EqualValues("Implementation.sol", sources[0].Name)
	s.Require().EqualValues("GPL-3.0", sources[0].License)
	s.Require().Len(sources[0].Urls, 2)
}

func (s *StorageTestSuite) TestSourceFilterDefaultLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Limit=0 should use default (10) via limitScope
	sources, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 3,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(sources, 2)
}

func (s *StorageTestSuite) TestSourceFilterCursorPagination() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Page 1: get first item
	page1, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 4,
		Limit:      1,
		Sort:       sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(page1, 1)
	s.Require().EqualValues(3, page1[0].Id)

	// Page 2: use cursor from last item of page 1
	page2, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 4,
		Limit:      1,
		Sort:       sdk.SortOrderAsc,
		CursorID:   page1[0].Id,
	})
	s.Require().NoError(err)
	s.Require().Len(page2, 1)
	s.Require().EqualValues(4, page2[0].Id)

	// Page 3: no more results
	page3, err := s.storage.Sources.Filter(ctx, storage.SourceListFilter{
		ContractId: 4,
		Limit:      1,
		Sort:       sdk.SortOrderAsc,
		CursorID:   page2[0].Id,
	})
	s.Require().NoError(err)
	s.Require().Len(page3, 0)
}

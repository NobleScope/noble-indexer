package postgres

import (
	"context"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestLogFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 8)

	// Check first log
	s.Require().EqualValues(1, logs[0].Id)
	s.Require().EqualValues(100, logs[0].Height)
	s.Require().EqualValues(0, logs[0].Index)
	s.Require().EqualValues("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", logs[0].Name)
	s.Require().EqualValues(1, logs[0].TxId)
	s.Require().EqualValues(1, logs[0].AddressId)
	s.Require().False(logs[0].Removed)

	// Check that topics are loaded
	s.Require().NotNil(logs[0].Topics)
	s.Require().Greater(len(logs[0].Topics), 0)

	// Check that relations are loaded with correct data
	s.Require().NotNil(logs[0].Tx.Hash)
	s.Require().NotNil(logs[0].Address.Hash)
	s.Require().EqualValues("0x90f5df4e03620cc55d3ea295bf8826f84465065340cb6d0d095166dd2465f283", logs[0].Tx.Hash.Hex())
	s.Require().EqualValues("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2", logs[0].Address.Hash.Hex())
}

func (s *StorageTestSuite) TestLogFilterDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 8)

	// Check descending order (by time and id)
	s.Require().EqualValues(8, logs[0].Id)
	s.Require().EqualValues(300, logs[0].Height)
	s.Require().EqualValues(1, logs[7].Id)
	s.Require().EqualValues(100, logs[7].Height)
}

func (s *StorageTestSuite) TestLogFilterByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(1)
	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		TxId:   &txId,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	// Check that all logs belong to tx_id 1
	for _, log := range logs {
		s.Require().EqualValues(1, log.TxId)
	}

	// Check first log details
	s.Require().EqualValues(1, logs[0].Id)
	s.Require().EqualValues("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", logs[0].Name)
	s.Require().EqualValues(0, logs[0].Index)

	// Check topics
	s.Require().NotNil(logs[0].Topics)
	s.Require().Greater(len(logs[0].Topics), 0)
}

func (s *StorageTestSuite) TestLogFilterByAddressId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(4)
	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		AddressId: &addressId,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	// Check that all logs belong to address_id 4
	for _, log := range logs {
		s.Require().EqualValues(4, log.AddressId)
	}

	// Check log details
	s.Require().EqualValues(4, logs[0].Id)
	s.Require().EqualValues(300, logs[0].Height)

	// Check topics are loaded
	s.Require().NotNil(logs[0].Topics)
	s.Require().Greater(len(logs[0].Topics), 0)
}

func (s *StorageTestSuite) TestLogFilterByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(300)
	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: &height,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 5)

	// Check that all logs are at height 300
	for _, log := range logs {
		s.Require().EqualValues(300, log.Height)
	}

	// Check log details
	s.Require().EqualValues(4, logs[0].Id)
	s.Require().EqualValues("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", logs[0].Name)

	// Check topics are loaded
	s.Require().NotNil(logs[0].Topics)
	s.Require().Greater(len(logs[0].Topics), 0)
}

func (s *StorageTestSuite) TestLogFilterByTimeRange() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:02Z")

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:    10,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 4)

	// Check that all logs are in the time range
	for _, log := range logs {
		s.Require().True(log.Time.Equal(timeFrom) || log.Time.After(timeFrom))
		s.Require().True(log.Time.Before(timeTo))
	}

	// Check heights (should be 200 and 300)
	s.Require().EqualValues(200, logs[0].Height)
	s.Require().EqualValues(200, logs[1].Height)
	s.Require().EqualValues(300, logs[2].Height)
	s.Require().EqualValues(300, logs[3].Height)
}

func (s *StorageTestSuite) TestLogFilterLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  2,
		Offset: 2,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 2)

	// With offset=2, should skip first two logs and return third and fourth
	s.Require().EqualValues(3, logs[0].Id)
	s.Require().EqualValues(200, logs[0].Height)
	s.Require().EqualValues(4, logs[1].Id)
	s.Require().EqualValues(300, logs[1].Height)
}

func (s *StorageTestSuite) TestLogFilterEmptyResult() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 100,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 0)
}

func (s *StorageTestSuite) TestLogFilterByNonExistentTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(999)
	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		TxId:   &txId,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 0)
}

func (s *StorageTestSuite) TestLogFilterCombinedFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(100)
	txId := uint64(1)

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: &height,
		TxId:   &txId,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	// Check that all logs match both filters
	for _, log := range logs {
		s.Require().EqualValues(100, log.Height)
		s.Require().EqualValues(1, log.TxId)
	}
}

func (s *StorageTestSuite) TestLogFilterLimitExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Request 100 logs when only 8 exist
	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  100,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 8)

	s.Require().EqualValues(100, logs[0].Height)
	s.Require().EqualValues(300, logs[7].Height)
}

func (s *StorageTestSuite) TestLogFilterDescWithTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(3)
	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
		TxId:   &txId,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 5)

	// Check descending order for tx_id 3 logs
	s.Require().EqualValues(8, logs[0].Id)
	s.Require().EqualValues(4, logs[4].Id)

	// Verify JOIN fields are loaded
	s.Require().NotNil(logs[0].Tx.Hash)
	s.Require().NotNil(logs[0].Address.Hash)
}

func (s *StorageTestSuite) TestLogFilterHeightAndAddressId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(300)
	addressId := uint64(4)

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		Height:    &height,
		AddressId: &addressId,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	// Check that log matches both filters
	s.Require().EqualValues(300, logs[0].Height)
	s.Require().EqualValues(4, logs[0].AddressId)
	s.Require().EqualValues(4, logs[0].Id)
}

func (s *StorageTestSuite) TestLogFilterTimeFromOnly() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:    10,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
		TimeFrom: timeFrom,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 7)

	// All logs should be >= timeFrom
	for _, log := range logs {
		s.Require().True(log.Time.Equal(timeFrom) || log.Time.After(timeFrom))
	}

	// First log should be at height 200, last at 300
	s.Require().EqualValues(200, logs[0].Height)
	s.Require().EqualValues(300, logs[6].Height)
}

func (s *StorageTestSuite) TestLogFilterTimeToOnly() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeTo, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:02Z")

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		TimeTo: timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 3)

	// All logs should be < timeTo
	for _, log := range logs {
		s.Require().True(log.Time.Before(timeTo))
	}

	// Should include heights 100 and 200
	s.Require().EqualValues(100, logs[0].Height)
	s.Require().EqualValues(200, logs[2].Height)
}

func (s *StorageTestSuite) TestLogFilterMinimalLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  1,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	// Should return only the first log
	s.Require().EqualValues(1, logs[0].Id)
	s.Require().EqualValues(100, logs[0].Height)

	// Verify JOIN fields work with limit=1
	s.Require().NotNil(logs[0].Tx.Hash)
	s.Require().NotNil(logs[0].Address.Hash)
}

func (s *StorageTestSuite) TestLogFilterExactBoundary() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:  10,
		Offset: 7,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	// Should return only the last log
	s.Require().EqualValues(8, logs[0].Id)
	s.Require().EqualValues(300, logs[0].Height)
}

func (s *StorageTestSuite) TestLogFilterAllFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(300)
	txId := uint64(3)
	addressId := uint64(4)
	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:01Z")

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		Height:    &height,
		TxId:      &txId,
		AddressId: &addressId,
		TimeFrom:  timeFrom,
		TimeTo:    timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	// Verify all filters are applied
	log := logs[0]
	s.Require().EqualValues(300, log.Height)
	s.Require().EqualValues(3, log.TxId)
	s.Require().EqualValues(4, log.AddressId)
	s.Require().True(log.Time.Equal(timeFrom) || log.Time.After(timeFrom))
	s.Require().True(log.Time.Before(timeTo))

	// Verify specific log
	s.Require().EqualValues(4, log.Id)
}

func (s *StorageTestSuite) TestLogFilterDescWithAddressId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addressId := uint64(1)
	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderDesc,
		AddressId: &addressId,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 2)

	// Check descending order
	s.Require().EqualValues(8, logs[0].Id)
	s.Require().EqualValues(1, logs[1].Id)

	// Both should have address_id = 1
	for _, log := range logs {
		s.Require().EqualValues(1, log.AddressId)
	}

	// Verify JOIN fields
	s.Require().NotNil(logs[0].Tx.Hash)
	s.Require().NotNil(logs[0].Address.Hash)
	s.Require().EqualValues("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2", logs[0].Address.Hash.Hex())
}

func (s *StorageTestSuite) TestLogFilterHeightAndTimeRange() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(300)
	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:03Z")

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:    10,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
		Height:   &height,
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 3)

	// All should be at height 300 and within time range
	for _, log := range logs {
		s.Require().EqualValues(300, log.Height)
		s.Require().True(log.Time.Equal(timeFrom) || log.Time.After(timeFrom))
		s.Require().True(log.Time.Before(timeTo))
	}
}

func (s *StorageTestSuite) TestLogFilterWithABI() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	logs, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		WithABI:   true,
		AddressId: uint64Ptr(3),
	})
	s.Require().NoError(err)
	s.Require().Len(logs, 1)

	for _, log := range logs {
		s.Require().NotNil(log.ContractABI)
	}
}

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestBlockLast() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	block, err := s.storage.Blocks.Last(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(5, block.Id)
	s.Require().EqualValues(500, block.Height)
	s.Require().EqualValues("30000000", block.GasLimit.String())
	s.Require().EqualValues("105000", block.GasUsed.String())
	s.Require().EqualValues(1400000000, block.BaseFeePerGas)
	s.Require().EqualValues(2, block.MinerId)
	s.Require().EqualValues("0x5234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", block.Hash.Hex())
}

func (s *StorageTestSuite) TestBlockByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	block, err := s.storage.Blocks.ByHeight(ctx, 200, false)
	s.Require().NoError(err)
	s.Require().EqualValues(2, block.Id)
	s.Require().EqualValues(200, block.Height)
	s.Require().EqualValues("30000000", block.GasLimit.String())
	s.Require().EqualValues("42000", block.GasUsed.String())
	s.Require().EqualValues(1100000000, block.BaseFeePerGas)
	s.Require().EqualValues(2, block.MinerId)
	s.Require().EqualValues("0x2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", block.Hash.Hex())
	s.Require().Nil(block.Stats)
}

func (s *StorageTestSuite) TestBlockByHeightWithStats() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	block, err := s.storage.Blocks.ByHeight(ctx, 200, true)
	s.Require().NoError(err)
	s.Require().EqualValues(2, block.Id)
	s.Require().EqualValues(200, block.Height)
	s.Require().EqualValues("30000000", block.GasLimit.String())
	s.Require().EqualValues("42000", block.GasUsed.String())
	s.Require().EqualValues(1100000000, block.BaseFeePerGas)
	s.Require().EqualValues(2, block.MinerId)
	s.Require().EqualValues("0x2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", block.Hash.Hex())

	// Check stats
	s.Require().NotNil(block.Stats)
	s.Require().EqualValues(2, block.Stats.Id)
	s.Require().EqualValues(200, block.Stats.Height)
	s.Require().EqualValues(2, block.Stats.TxCount)
	s.Require().EqualValues(11000, block.Stats.BlockTime)
}

func (s *StorageTestSuite) TestBlockByHeightWithoutStats() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Block 400 exists but has no stats
	block, err := s.storage.Blocks.ByHeight(ctx, 400, true)
	s.Require().NoError(err)
	s.Require().EqualValues(4, block.Id)
	s.Require().EqualValues(400, block.Height)
	s.Require().Nil(block.Stats)
}

func (s *StorageTestSuite) TestBlockByHeightNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.storage.Blocks.ByHeight(ctx, types.Level(99999), false)
	s.Require().Error(err)
	s.Require().ErrorIs(err, sql.ErrNoRows)
}

func (s *StorageTestSuite) TestBlockListWithStats() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	blocks, err := s.storage.Blocks.Filter(ctx, storage.BlockListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		WithStats: true,
	})
	s.Require().NoError(err)
	s.Require().Len(blocks, 5)

	// Check first block
	s.Require().EqualValues(1, blocks[0].Id)
	s.Require().EqualValues(100, blocks[0].Height)
	s.Require().NotNil(blocks[0].Stats)
	s.Require().EqualValues(1, blocks[0].Stats.TxCount)

	// Check second block
	s.Require().EqualValues(2, blocks[1].Id)
	s.Require().EqualValues(200, blocks[1].Height)
	s.Require().NotNil(blocks[1].Stats)
	s.Require().EqualValues(2, blocks[1].Stats.TxCount)

	// Check last block
	s.Require().EqualValues(5, blocks[4].Id)
	s.Require().EqualValues(500, blocks[4].Height)
	s.Require().NotNil(blocks[4].Stats)
	s.Require().EqualValues(5, blocks[4].Stats.TxCount)
}

func (s *StorageTestSuite) TestBlockListWithStatsDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	blocks, err := s.storage.Blocks.Filter(ctx, storage.BlockListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderDesc,
		WithStats: true,
	})
	s.Require().NoError(err)
	s.Require().Len(blocks, 5)

	// Check descending order
	s.Require().EqualValues(5, blocks[0].Id)
	s.Require().EqualValues(500, blocks[0].Height)
	s.Require().EqualValues(1, blocks[4].Id)
	s.Require().EqualValues(100, blocks[4].Height)
}

func (s *StorageTestSuite) TestBlockListWithStatsLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	blocks, err := s.storage.Blocks.Filter(ctx, storage.BlockListFilter{
		Limit:     2,
		Offset:    1,
		Sort:      sdk.SortOrderAsc,
		WithStats: true,
	})
	s.Require().NoError(err)
	s.Require().Len(blocks, 2)

	// With offset=1, should skip first block and return second and third
	s.Require().EqualValues(2, blocks[0].Id)
	s.Require().EqualValues(200, blocks[0].Height)
	s.Require().EqualValues(3, blocks[1].Id)
	s.Require().EqualValues(300, blocks[1].Height)
}

func (s *StorageTestSuite) TestBlockListWithStatsEmptyResult() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	blocks, err := s.storage.Blocks.Filter(ctx, storage.BlockListFilter{
		Limit:     10,
		Offset:    100,
		Sort:      sdk.SortOrderAsc,
		WithStats: true,
	})
	s.Require().NoError(err)
	s.Require().Len(blocks, 0)
}

func (s *StorageTestSuite) TestBlockListWithStatsBlockWithoutStats() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	blocks, err := s.storage.Blocks.Filter(ctx, storage.BlockListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		WithStats: true,
	})
	s.Require().NoError(err)
	s.Require().Len(blocks, 5)

	// Block 4 (height 400) has no stats
	s.Require().EqualValues(4, blocks[3].Id)
	s.Require().EqualValues(400, blocks[3].Height)
	s.Require().Nil(blocks[3].Stats)
}

func (s *StorageTestSuite) TestBlockFilterWithoutStats() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	blocks, err := s.storage.Blocks.Filter(ctx, storage.BlockListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		WithStats: false,
	})
	s.Require().NoError(err)
	s.Require().Len(blocks, 5)

	// Check that Stats are not loaded
	for _, block := range blocks {
		s.Require().Nil(block.Stats)
	}

	// Check basic block data is present
	s.Require().EqualValues(1, blocks[0].Id)
	s.Require().EqualValues(100, blocks[0].Height)
}

func (s *StorageTestSuite) TestBlockFilterLimitExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Request 100 blocks when only 5 exist
	blocks, err := s.storage.Blocks.Filter(ctx, storage.BlockListFilter{
		Limit:     100,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		WithStats: false,
	})
	s.Require().NoError(err)
	s.Require().Len(blocks, 5)

	// Verify all 5 blocks are returned
	s.Require().EqualValues(1, blocks[0].Id)
	s.Require().EqualValues(5, blocks[4].Id)
}
package postgres

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestBlockStatsByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	block, err := s.storage.BlockStats.ByHeight(ctx, 200)
	s.Require().NoError(err)
	s.Require().EqualValues(2, block.Id)
	s.Require().EqualValues(200, block.Height)
	s.Require().EqualValues(2, block.TxCount)
	s.Require().EqualValues(11000, block.BlockTime)
}

func (s *StorageTestSuite) TestBlockStatsAvgBlockTime() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	blockTime, err := s.storage.BlockStats.AvgBlockTime(ctx, from)
	s.Require().NoError(err)
	s.Require().EqualValues(11500, blockTime)
}

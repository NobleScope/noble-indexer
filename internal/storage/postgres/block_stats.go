package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type BlockStats struct {
	*postgres.Table[*storage.BlockStats]
}

// NewBlockStats -
func NewBlockStats(db *database.Bun) *BlockStats {
	return &BlockStats{
		Table: postgres.NewTable[*storage.BlockStats](db),
	}
}

// ByHeight -
func (b *BlockStats) ByHeight(ctx context.Context, height pkgTypes.Level) (stats storage.BlockStats, err error) {
	err = b.DB().NewSelect().Model(&stats).
		Where("height = ?", height).
		Limit(1).
		Scan(ctx)

	return
}

// AvgBlockTime -
func (b *BlockStats) AvgBlockTime(ctx context.Context, from time.Time) (blockTime float64, err error) {
	err = b.DB().NewSelect().
		Model((*storage.BlockStats)(nil)).
		ColumnExpr("AVG(block_time)").
		Where("time >= ?", from).
		Scan(ctx, &blockTime)
	return
}

package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/go-lib/database"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type Block struct {
	*postgres.Table[*storage.Block]
}

// NewBlock -
func NewBlock(db *database.Bun) *Block {
	return &Block{
		Table: postgres.NewTable[*storage.Block](db),
	}
}

// ByHeight -
func (b *Block) ByHeight(ctx context.Context, height types.Level, withStats bool) (block storage.Block, err error) {
	if !withStats {
		err = b.DB().NewSelect().Model(&block).
			Where("block.height = ?", height).
			Limit(1).
			Scan(ctx)
		return
	}

	subQuery := b.DB().NewSelect().Model(&block).
		Where("block.height = ?", height).
		Limit(1)

	err = b.DB().NewSelect().
		ColumnExpr("block.*").
		ColumnExpr("stats.id AS stats__id, stats.height AS stats__height, stats.time AS stats__time, stats.tx_count AS stats__tx_count, stats.block_time AS stats__block_time").
		TableExpr("(?) as block", subQuery).
		Join("LEFT JOIN block_stats AS stats ON (stats.height = block.height) AND (stats.time = block.time)").
		Scan(ctx, &block)

	return
}

// ListWithStats -
func (b *Block) ListWithStats(ctx context.Context, limit, offset uint64, order sdk.SortOrder) (blocks []*storage.Block, err error) {
	subQuery := b.DB().NewSelect().
		Model(&blocks)

	//nolint:gosec
	subQuery = limitScope(subQuery, int(limit))
	if offset > 0 {
		//nolint:gosec
		subQuery = subQuery.Offset(int(offset))
	}

	query := b.DB().NewSelect().
		ColumnExpr("block.*").
		ColumnExpr("stats.id AS stats__id, stats.height AS stats__height, stats.time AS stats__time, stats.tx_count AS stats__tx_count, stats.block_time AS stats__block_time").
		TableExpr("(?) as block", subQuery).
		Join("LEFT JOIN block_stats as stats ON (stats.height = block.height) AND (stats.time = block.time)")
	query = sortScope(query, "block.height", order)

	err = query.Scan(ctx, &blocks)
	return
}

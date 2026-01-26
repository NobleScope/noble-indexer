package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/go-lib/database"
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

// Last -
func (b *Block) Last(ctx context.Context) (block storage.Block, err error) {
	err = b.DB().NewSelect().Model(&block).
		Order("id DESC").
		Limit(1).
		Scan(ctx)
	return
}

// ByHeight -
func (b *Block) ByHeight(ctx context.Context, height types.Level, withStats bool) (block storage.Block, err error) {
	query := b.DB().NewSelect().
		Model(&block).
		ColumnExpr("block.*").
		ColumnExpr("address.hash AS miner__hash").
		Join("LEFT JOIN address ON address.id = block.miner_id").
		Where("block.height = ?", height).
		Limit(1)

	if withStats {
		query = query.
			ColumnExpr("stats.id AS stats__id, stats.height AS stats__height, stats.time AS stats__time, stats.tx_count AS stats__tx_count, stats.block_time AS stats__block_time").
			Join("LEFT JOIN block_stats AS stats ON (stats.height = block.height) AND (stats.time = block.time)")
	}

	err = query.Scan(ctx, &block)
	return
}

// Filter -
func (b *Block) Filter(ctx context.Context, fltrs storage.BlockListFilter) (blocks []storage.Block, err error) {
	query := b.DB().NewSelect().
		Model(&blocks)

	query = sortTimeIDScope(query, fltrs.Sort)
	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)

	query = b.DB().NewSelect().
		ColumnExpr("block.*").
		ColumnExpr("address.hash AS miner__hash").
		TableExpr("(?) as block", query).
		Join("LEFT JOIN address ON address.id = block.miner_id")

	if fltrs.WithStats {
		query = query.
			ColumnExpr("stats.id AS stats__id, stats.height AS stats__height, stats.time AS stats__time, stats.tx_count AS stats__tx_count, stats.block_time AS stats__block_time").
			Join("LEFT JOIN block_stats as stats ON (stats.height = block.height) AND (stats.time = block.time)")
	}

	query = sortTimeIDScope(query, fltrs.Sort)
	err = query.Scan(ctx, &blocks)

	return
}

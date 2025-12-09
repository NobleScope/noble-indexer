package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// Contract -
type Contract struct {
	*postgres.Table[*storage.Contract]
}

// NewContract -
func NewContract(db *database.Bun) *Contract {
	return &Contract{
		Table: postgres.NewTable[*storage.Contract](db),
	}
}

// PendingMetadata -
func (c *Contract) PendingMetadata(ctx context.Context, retryDelay time.Duration, limit int) (contracts []*storage.Contract, err error) {
	threshold := time.Now().UTC().Add(-retryDelay)
	err = c.DB().NewSelect().
		Model(&contracts).
		Relation("Address").
		Where("metadata_link IS NOT NULL AND metadata_link <> ''").
		Where("status = 'pending' AND (updated_at < ? OR retry_count = 0)", threshold).
		Order("id ASC").
		Limit(limit).
		Scan(ctx)

	return
}

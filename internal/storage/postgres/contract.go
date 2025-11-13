package postgres

import (
	"context"

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

// ListWithMetadata -
func (c *Contract) ListWithMetadata(ctx context.Context, startId uint64) (contracts []*storage.Contract, err error) {
	err = c.DB().NewSelect().
		Model(&contracts).
		Where("id > ?", startId).
		Where("metadata_link IS NOT NULL AND metadata_link <> ''").
		Order("id ASC").
		Scan(ctx)

	return
}

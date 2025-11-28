package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type Token struct {
	*postgres.Table[*storage.Token]
}

// NewToken -
func NewToken(db *database.Bun) *Token {
	return &Token{
		Table: postgres.NewTable[*storage.Token](db),
	}
}

// PendingMetadata -
func (c *Token) PendingMetadata(ctx context.Context, retryDelay time.Duration, limit int) (tokens []*storage.Token, err error) {
	threshold := time.Now().UTC().Add(-retryDelay)

	err = c.DB().NewSelect().
		Model(&tokens).
		Where("status = 'pending' AND (updated_at < ? OR retry_count = 0)", threshold).
		Order("id ASC").
		Limit(limit).
		Scan(ctx)

	return
}

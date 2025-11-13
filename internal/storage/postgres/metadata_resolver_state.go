package postgres

import (
	"context"
	"database/sql"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
	"github.com/pkg/errors"
)

// MetadataResolverState -
type MetadataResolverState struct {
	*postgres.Table[*storage.MetadataResolverState]
}

// NewMetadataResolverState -
func NewMetadataResolverState(db *database.Bun) *MetadataResolverState {
	return &MetadataResolverState{
		Table: postgres.NewTable[*storage.MetadataResolverState](db),
	}
}

// ByName -
func (s *MetadataResolverState) ByName(ctx context.Context, name string) (state storage.MetadataResolverState, err error) {
	err = s.DB().NewSelect().Model(&state).
		Where("name = ?", name).
		Scan(ctx)
	return
}

func (s *MetadataResolverState) IsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

package postgres

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// Source -
type Source struct {
	*postgres.Table[*storage.Source]
}

// NewSource -
func NewSource(db *database.Bun) *Source {
	return &Source{
		Table: postgres.NewTable[*storage.Source](db),
	}
}

// ByContractId -
func (s *Source) ByContractId(ctx context.Context, id uint64, limit, offset int) (sources []storage.Source, err error) {
	err = s.DB().NewSelect().
		Model((*storage.Source)(nil)).
		Where("contract_id = ?", id).
		Limit(limit).
		Offset(offset).
		Scan(ctx, &sources)

	return
}

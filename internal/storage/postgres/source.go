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

// Filter -
func (s *Source) Filter(ctx context.Context, filter storage.SourceListFilter) (sources []storage.Source, err error) {
	query := s.DB().NewSelect().
		Model((*storage.Source)(nil)).
		Where("contract_id = ?", filter.ContractId)

	if filter.CursorID > 0 {
		query = cursorIDScope(query, filter.Sort, filter.CursorID)
	} else {
		query = query.Offset(filter.Offset)
	}

	query = limitScope(query, filter.Limit)
	query = sortScope(query, "id", filter.Sort)

	err = query.Scan(ctx, &sources)
	return
}

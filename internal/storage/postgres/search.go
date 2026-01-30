package postgres

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
)

// Search -
type Search struct {
	db *database.Bun
}

// NewSearch -
func NewSearch(db *database.Bun) *Search {
	return &Search{
		db: db,
	}
}

// Search -
func (s *Search) Search(ctx context.Context, query []byte, limit, offset int) (results []storage.SearchResult, err error) {
	blockQuery := s.db.DB().NewSelect().
		Model((*storage.Block)(nil)).
		ColumnExpr("id, ? as value, 'block' as type", hex.EncodeToString(query)).
		Where("hash = ?", query)
	txQuery := s.db.DB().NewSelect().
		Model((*storage.Tx)(nil)).
		ColumnExpr("id, hash as value, 'tx' as type").
		Where("hash = ?", query)

	union := blockQuery.UnionAll(txQuery)

	q := s.db.DB().NewSelect().TableExpr("(?) as search", union)

	q = limitScope(q, limit)
	if offset > 0 {
		q = q.Offset(offset)
	}
	err = q.Scan(ctx, &results)

	return
}

// SearchText -
func (s *Search) SearchText(ctx context.Context, text string, limit, offset int) (results []storage.SearchResult, err error) {
	text = strings.ToUpper(text)
	text = "%" + text + "%"
	tokenNameQuery := s.db.DB().NewSelect().
		Model((*storage.Token)(nil)).
		ColumnExpr("id, name as value, 'token' as type").
		Where("name ILIKE ?", text)
	tokenSymbolQuery := s.db.DB().NewSelect().
		Model((*storage.Token)(nil)).
		ColumnExpr("id, symbol as value, 'token' as type").
		Where("symbol ILIKE ?", text)

	union := tokenNameQuery.UnionAll(tokenSymbolQuery)

	q := s.db.DB().NewSelect().
		TableExpr("(?) as search", union)
	q = limitScope(q, limit)
	if offset > 0 {
		q = q.Offset(offset)
	}
	err = q.Scan(ctx, &results)

	return
}

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

func (s *Search) Search(ctx context.Context, query []byte) (results []storage.SearchResult, err error) {
	blockQuery := s.db.DB().NewSelect().
		Model((*storage.Block)(nil)).
		ColumnExpr("id, ? as value, 'block' as type", hex.EncodeToString(query)).
		Where("hash = ?", query)
	txQuery := s.db.DB().NewSelect().
		Model((*storage.Tx)(nil)).
		ColumnExpr("id, hash as value, 'tx' as type").
		Where("hash = ?", query)

	union := blockQuery.UnionAll(txQuery)

	err = s.db.DB().NewSelect().
		TableExpr("(?) as search", union).
		Limit(10).
		Offset(0).
		Scan(ctx, &results)

	return
}

func (s *Search) SearchText(ctx context.Context, text string) (results []storage.SearchResult, err error) {
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

	err = s.db.DB().NewSelect().
		TableExpr("(?) as search", union).
		Limit(10).
		Offset(0).
		Scan(ctx, &results)

	return
}

package postgres

import (
	"context"
	"github.com/shopspring/decimal"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
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
func (t *Token) PendingMetadata(ctx context.Context, retryDelay time.Duration, limit int) (tokens []*storage.Token, err error) {
	threshold := time.Now().UTC().Add(-retryDelay)

	query := t.DB().NewSelect().
		Model((*storage.Token)(nil)).
		Where("status = 'pending' AND (updated_at < ? OR retry_count = 0)", threshold).
		Order("id ASC").
		Limit(limit)

	err = t.DB().NewSelect().TableExpr("(?) AS token", query).
		ColumnExpr("token.*").
		ColumnExpr("contract_address.hash AS contract__address__hash").
		Join("LEFT JOIN contract ON contract.id = token.contract_id").
		Join("LEFT JOIN address AS contract_address ON contract_address.id = contract.id").
		Scan(ctx, &tokens)

	return
}

// Filter -
func (t *Token) Filter(ctx context.Context, filter storage.TokenListFilter) (tokens []storage.Token, err error) {
	query := t.DB().NewSelect().
		Model(&tokens)

	query = tokenListFilter(query, filter)
	err = t.DB().NewSelect().TableExpr("(?) AS token", query).
		ColumnExpr("token.*").
		ColumnExpr("contract_address.hash AS contract__address__hash").
		Join("LEFT JOIN contract ON contract.id = token.contract_id").
		Join("LEFT JOIN address AS contract_address ON contract_address.id = contract.id").
		Scan(ctx, &tokens)

	return
}

// Get -
func (t *Token) Get(ctx context.Context, contractId uint64, tokenId decimal.Decimal) (token storage.Token, err error) {
	query := t.DB().NewSelect().
		Model(&token).
		Where("contract_id = ? AND token_id = ?", contractId, tokenId)

	err = t.DB().NewSelect().TableExpr("(?) AS token", query).
		ColumnExpr("token.*").
		ColumnExpr("contract_address.hash AS contract__address__hash").
		Join("LEFT JOIN contract ON contract.id = token.contract_id").
		Join("LEFT JOIN address AS contract_address ON contract_address.id = contract.id").
		Scan(ctx, &token)

	return
}

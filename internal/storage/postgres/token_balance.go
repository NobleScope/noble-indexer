package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type TokenBalance struct {
	*postgres.Table[*storage.TokenBalance]
}

// NewTokenBalance -
func NewTokenBalance(db *database.Bun) *TokenBalance {
	return &TokenBalance{
		Table: postgres.NewTable[*storage.TokenBalance](db),
	}
}

func (t *TokenBalance) Filter(ctx context.Context, filter storage.TokenBalanceListFilter) (tb []storage.TokenBalance, err error) {
	query := t.DB().NewSelect().
		Model(&tb)

	query = tokenBalanceListFilter(query, filter)
	err = t.DB().NewSelect().TableExpr("(?) AS token_balance", query).
		ColumnExpr("token_balance.*").
		ColumnExpr("contract.address AS contract__address").
		ColumnExpr("address.address AS address__address").
		Join("LEFT JOIN contract ON contract.id = token_balance.contract_id").
		Join("LEFT JOIN address ON address.id = token_balance.address_id").
		Scan(ctx, &tb)

	return
}

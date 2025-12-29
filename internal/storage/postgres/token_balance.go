package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
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

// Filter -
func (t *TokenBalance) Filter(ctx context.Context, filter storage.TokenBalanceListFilter) (tb []storage.TokenBalance, err error) {
	query := t.DB().NewSelect().
		Model(&tb)

	query = tokenBalanceListFilter(query, filter)

	outerQuery := t.DB().NewSelect().TableExpr("(?) AS token_balance", query).
		ColumnExpr("token_balance.*").
		ColumnExpr("contract_address.hash AS contract__address__hash").
		ColumnExpr("address.hash AS address__hash").
		Join("LEFT JOIN contract ON contract.id = token_balance.contract_id").
		Join("LEFT JOIN address AS contract_address ON contract_address.id = contract.id").
		Join("LEFT JOIN address ON address.id = token_balance.address_id")

	outerQuery = sortScope(outerQuery, "token_balance.balance", filter.Sort)
	outerQuery = sortScope(outerQuery, "token_balance.id", sdk.SortOrderAsc)

	err = outerQuery.Scan(ctx, &tb)

	return
}

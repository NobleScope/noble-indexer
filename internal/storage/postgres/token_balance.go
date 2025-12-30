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

	outerQuery = sortMultipleScope(outerQuery, []SortField{
		{Field: "token_balance.balance", Order: filter.Sort},
		{Field: "token_balance.id", Order: filter.Sort},
	})

	err = outerQuery.Scan(ctx, &tb)

	return
}

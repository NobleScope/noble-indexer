package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// Address -
type Address struct {
	*postgres.Table[*storage.Address]
}

// NewAddress -
func NewAddress(db *database.Bun) *Address {
	return &Address{
		Table: postgres.NewTable[*storage.Address](db),
	}
}

func (a *Address) ListWithBalance(ctx context.Context, filters storage.AddressListFilter) (result []storage.Address, err error) {
	if filters.SortField == "value" {
		query := a.DB().NewSelect().
			Model((*storage.Balance)(nil))

		query = addressListFilter(query, filters)
		query = a.DB().NewSelect().
			TableExpr("(?) as balance", query).
			ColumnExpr("address.*").
			ColumnExpr("balance.currency as balance__currency, balance.value as balance__value").
			Join("left join address on balance.id = address.id")

		query = addressListFilter(query, filters)
		err = query.Scan(ctx, &result)
	} else {
		query := a.DB().NewSelect().
			Model((*storage.Address)(nil))

		query = addressListFilter(query, filters)
		query = a.DB().NewSelect().
			TableExpr("(?) as address", query).
			ColumnExpr("address.*").
			ColumnExpr("balance.currency as balance__currency, balance.value as balance__value").
			Join("left join balance on balance.id = address.id")

		query = addressListFilter(query, filters)
		err = query.Scan(ctx, &result)
	}

	return
}

func (a *Address) ByHash(ctx context.Context, hash string) (address storage.Address, err error) {
	addressQuery := a.DB().NewSelect().
		Model((*storage.Address)(nil)).
		Where("address = ?", hash)

	err = a.DB().NewSelect().TableExpr("(?) as address", addressQuery).
		ColumnExpr("address.*").
		ColumnExpr("balance.currency as balance__currency, balance.value as balance__value").
		Join("left join balance on balance.id = address.id").
		Scan(ctx, &address)
	return
}

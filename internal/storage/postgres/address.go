package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
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

// ListWithBalance -
func (a *Address) ListWithBalance(ctx context.Context, filters storage.AddressListFilter) (result []storage.Address, err error) {
	if filters.SortField == "value" {
		query := a.DB().NewSelect().
			Model((*storage.Balance)(nil))

		query = addressListFilter(query, filters)
		query = a.DB().NewSelect().
			TableExpr("(?) AS balance", query).
			ColumnExpr("address.*").
			ColumnExpr("balance.currency AS balance__currency, balance.value AS balance__value").
			Join("LEFT JOIN address ON balance.id = address.id")

		query = addressListFilter(query, filters)
		err = query.Scan(ctx, &result)
	} else {
		query := a.DB().NewSelect().
			Model((*storage.Address)(nil))

		query = addressListFilter(query, filters)
		query = a.DB().NewSelect().
			TableExpr("(?) AS address", query).
			ColumnExpr("address.*").
			ColumnExpr("balance.currency AS balance__currency, balance.value AS balance__value").
			Join("LEFT JOIN balance ON balance.id = address.id")

		query = addressListFilter(query, filters)
		err = query.Scan(ctx, &result)
	}

	return
}

// ByHash -
func (a *Address) ByHash(ctx context.Context, hash pkgTypes.Hex) (address storage.Address, err error) {
	addressQuery := a.DB().NewSelect().
		Model((*storage.Address)(nil)).
		Where("address = ?", hash.String())

	err = a.DB().NewSelect().TableExpr("(?) AS address", addressQuery).
		ColumnExpr("address.*").
		ColumnExpr("balance.currency AS balance__currency, balance.value AS balance__value").
		Join("LEFT JOIN balance ON balance.id = address.id").
		Scan(ctx, &address)

	return
}

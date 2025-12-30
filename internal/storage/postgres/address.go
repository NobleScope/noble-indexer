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
		balanceQuery := a.DB().NewSelect().
			Model((*storage.Balance)(nil))

		balanceQuery = balanceListFilter(balanceQuery, filters)
		query := a.DB().NewSelect().
			TableExpr("(?) as balance", balanceQuery).
			ColumnExpr("address.*").
			ColumnExpr("balance.id AS balance__id, balance.value AS balance__value").
			Join("LEFT JOIN address ON address.id = balance.id")

		if filters.OnlyContracts {
			query = query.Where("address.is_contract = ?", true)
		}

		err = query.Scan(ctx, &result)
	} else {
		addressQuery := a.DB().NewSelect().
			Model((*storage.Address)(nil))

		addressQuery = addressListFilter(addressQuery, filters)
		err = a.DB().NewSelect().
			TableExpr("(?) as address", addressQuery).
			ColumnExpr("address.*").
			ColumnExpr("balance.id AS balance__id, balance.value AS balance__value").
			Join("LEFT JOIN balance ON balance.id = address.id").
			Scan(ctx, &result)
	}

	return
}

// ByHash -
func (a *Address) ByHash(ctx context.Context, hash pkgTypes.Hex) (address storage.Address, err error) {
	addressQuery := a.DB().NewSelect().
		Model((*storage.Address)(nil)).
		Where("hash = ?", hash)

	err = a.DB().NewSelect().TableExpr("(?) AS address", addressQuery).
		ColumnExpr("address.*").
		ColumnExpr("balance.id AS balance__id, balance.value AS balance__value").
		Join("LEFT JOIN balance ON balance.id = address.id").
		Scan(ctx, &address)

	return
}

package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/go-lib/database"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type Tx struct {
	*postgres.Table[*storage.Tx]
}

// NewTx -
func NewTx(db *database.Bun) *Tx {
	return &Tx{
		Table: postgres.NewTable[*storage.Tx](db),
	}
}

// ByHeight -
func (t *Tx) ByHeight(
	ctx context.Context,
	height pkgTypes.Level,
	limit, offset int,
	order sdk.SortOrder,
) (txs []*storage.Tx, err error) {
	subQuery := t.DB().NewSelect().
		Model(&txs).
		Where("tx.height = ?", height)

	subQuery = sortScope(subQuery, "id", order)

	subQuery = limitScope(subQuery, limit)
	if offset > 0 {
		subQuery = subQuery.Offset(offset)
	}

	err = t.DB().NewSelect().
		ColumnExpr("tx.*").
		ColumnExpr("from_addr.id AS from_address__id, from_addr.first_height AS from_address__first_height, from_addr.last_height AS from_address__last_height, from_addr.hash AS from_address__hash, from_addr.is_contract AS from_address__is_contract").
		ColumnExpr("to_addr.id AS to_address__id, to_addr.first_height AS to_address__first_height, to_addr.last_height AS to_address__last_height, to_addr.hash AS to_address__hash, to_addr.is_contract AS to_address__is_contract").
		TableExpr("(?) AS tx", subQuery).
		Join("INNER JOIN address AS from_addr ON from_addr.id = tx.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = tx.to_address_id").
		Scan(ctx, &txs)

	return
}

// ByHash -
func (t *Tx) ByHash(ctx context.Context, hash pkgTypes.Hex) (tx storage.Tx, err error) {
	subQuery := t.DB().NewSelect().
		Model(&tx).
		Where("tx.hash = ?", hash)

	err = t.DB().NewSelect().
		ColumnExpr("tx.*").
		ColumnExpr("from_addr.id AS from_address__id, from_addr.first_height AS from_address__first_height, from_addr.last_height AS from_address__last_height, from_addr.hash AS from_address__hash, from_addr.is_contract AS from_address__is_contract").
		ColumnExpr("to_addr.id AS to_address__id, to_addr.first_height AS to_address__first_height, to_addr.last_height AS to_address__last_height, to_addr.hash AS to_address__hash, to_addr.is_contract AS to_address__is_contract").
		TableExpr("(?) AS tx", subQuery).
		Join("INNER JOIN address AS from_addr ON from_addr.id = tx.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = tx.to_address_id").
		Scan(ctx, &tx)

	return
}

// Filter -
func (t *Tx) Filter(ctx context.Context, filter storage.TxListFilter) (txs []storage.Tx, err error) {
	query := t.DB().NewSelect().
		Model(&txs)

	query = txListFilter(query, filter)

	outerQuery := t.DB().NewSelect().
		ColumnExpr("tx.*").
		ColumnExpr("from_addr.id AS from_address__id, from_addr.first_height AS from_address__first_height, from_addr.last_height AS from_address__last_height, from_addr.hash AS from_address__hash, from_addr.is_contract AS from_address__is_contract").
		ColumnExpr("to_addr.id AS to_address__id, to_addr.first_height AS to_address__first_height, to_addr.last_height AS to_address__last_height, to_addr.hash AS to_address__hash, to_addr.is_contract AS to_address__is_contract").
		TableExpr("(?) AS tx", query).
		Join("LEFT JOIN address AS from_addr ON from_addr.id = tx.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = tx.to_address_id")

	outerQuery = sortTimeIDScope(outerQuery, filter.Sort)
	err = outerQuery.Scan(ctx, &txs)

	return
}

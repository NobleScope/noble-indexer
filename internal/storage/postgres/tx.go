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
		ColumnExpr("from_addr.id AS from_address__id, from_addr.height AS from_address__height, from_addr.last_height AS from_address__last_height, from_addr.address AS from_address__address, from_addr.is_contract AS from_address__is_contract").
		ColumnExpr("to_addr.id AS to_address__id, to_addr.height AS to_address__height, to_addr.last_height AS to_address__last_height, to_addr.address AS to_address__address, to_addr.is_contract AS to_address__is_contract").
		ColumnExpr("contract.id AS contract__id, contract.address AS contract__address, contract.code AS contract__code, contract.verified AS contract__verified, contract.tx_id AS contract__tx_id").
		TableExpr("(?) AS tx", subQuery).
		Join("INNER JOIN address AS from_addr ON from_addr.id = tx.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = tx.to_address_id").
		Join("LEFT JOIN contract ON contract.id = tx.contract_id").
		Scan(ctx, &txs)

	return
}

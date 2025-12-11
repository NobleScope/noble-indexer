package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type Transfer struct {
	*postgres.Table[*storage.Transfer]
}

// NewTransfer -
func NewTransfer(db *database.Bun) *Transfer {
	return &Transfer{
		Table: postgres.NewTable[*storage.Transfer](db),
	}
}

func (t *Transfer) Filter(ctx context.Context, filter storage.TransferListFilter) (transfers []storage.Transfer, err error) {
	query := t.DB().NewSelect().
		Model(&transfers)

	query = transferListFilter(query, filter)
	err = t.DB().NewSelect().
		ColumnExpr("transfer.*").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("from_addr.address AS from_address__address").
		ColumnExpr("to_addr.address AS to_address__address").
		ColumnExpr("contract.id AS contract__id, contract.address AS contract__address, contract.code AS contract__code, contract.verified as contract__verified").
		TableExpr("(?) AS transfer", query).
		Join("LEFT JOIN tx ON tx.id = transfer.tx_id").
		Join("LEFT JOIN address AS from_addr ON from_addr.id = transfer.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = transfer.to_address_id").
		Join("LEFT JOIN contract ON contract.id = transfer.contract_id").
		Scan(ctx, &transfers)

	return
}

func (t *Transfer) Get(ctx context.Context, id uint64) (transfer storage.Transfer, err error) {
	query := t.DB().NewSelect().
		Model(&transfer).
		Where("id = ?", id).
		Limit(1)

	err = t.DB().NewSelect().
		ColumnExpr("transfer.*").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("from_addr.address AS from_address__address").
		ColumnExpr("to_addr.address AS to_address__address").
		ColumnExpr("contract.id AS contract__id, contract.address AS contract__address, contract.code AS contract__code, contract.verified as contract__verified").
		TableExpr("(?) AS transfer", query).
		Join("LEFT JOIN tx ON tx.id = transfer.tx_id").
		Join("LEFT JOIN address AS from_addr ON from_addr.id = transfer.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = transfer.to_address_id").
		Join("LEFT JOIN contract ON contract.id = transfer.contract_id").
		Scan(ctx, &transfer)

	return
}

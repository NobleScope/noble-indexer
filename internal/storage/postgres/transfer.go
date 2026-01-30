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

// Filter -
func (t *Transfer) Filter(ctx context.Context, filter storage.TransferListFilter) (transfers []storage.Transfer, err error) {
	query := t.DB().NewSelect().
		Model(&transfers)

	query = transferListFilter(query, filter)

	outerQuery := t.DB().NewSelect().
		ColumnExpr("transfer.*").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("from_addr.hash AS from_address__hash").
		ColumnExpr("to_addr.hash AS to_address__hash").
		ColumnExpr("contract.id AS contract__id, contract_addr.hash AS contract__address__hash, contract.code AS contract__code, contract.verified as contract__verified").
		ColumnExpr("token.name AS token__name, token.symbol AS token__symbol, token.decimals AS token__decimals, token.type AS token__type, token.supply AS token__supply, token.transfers_count as token__transfers_count").
		TableExpr("(?) AS transfer", query).
		Join("LEFT JOIN tx ON tx.id = transfer.tx_id").
		Join("LEFT JOIN address AS from_addr ON from_addr.id = transfer.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = transfer.to_address_id").
		Join("LEFT JOIN contract ON contract.id = transfer.contract_id").
		Join("LEFT JOIN address AS contract_addr ON contract_addr.id = contract.id").
		Join("LEFT JOIN token ON token.token_id = transfer.token_id AND transfer.contract_id = token.contract_id")

	outerQuery = sortTimeIDScope(outerQuery, filter.Sort)
	err = outerQuery.Scan(ctx, &transfers)

	return
}

// Get -
func (t *Transfer) Get(ctx context.Context, id uint64) (transfer storage.Transfer, err error) {
	query := t.DB().NewSelect().
		Model(&transfer).
		Where("id = ?", id).
		Limit(1)

	err = t.DB().NewSelect().
		ColumnExpr("transfer.*").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("from_addr.hash AS from_address__hash").
		ColumnExpr("to_addr.hash AS to_address__hash").
		ColumnExpr("contract.id AS contract__id, contract_addr.hash AS contract__address__hash, contract.code AS contract__code, contract.verified as contract__verified").
		ColumnExpr("token.name AS token__name, token.symbol AS token__symbol, token.decimals AS token__decimals, token.type AS token__type, token.supply AS token__supply, token.transfers_count as token__transfers_count").
		TableExpr("(?) AS transfer", query).
		Join("LEFT JOIN tx ON tx.id = transfer.tx_id").
		Join("LEFT JOIN address AS from_addr ON from_addr.id = transfer.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = transfer.to_address_id").
		Join("LEFT JOIN contract ON contract.id = transfer.contract_id").
		Join("LEFT JOIN address AS contract_addr ON contract_addr.id = contract.id").
		Join("LEFT JOIN token on token.token_id = transfer.token_id AND transfer.contract_id = token.contract_id").
		Scan(ctx, &transfer)

	return
}

package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type Trace struct {
	*postgres.Table[*storage.Trace]
}

// NewTrace -
func NewTrace(db *database.Bun) *Trace {
	return &Trace{
		Table: postgres.NewTable[*storage.Trace](db),
	}
}

// Filter -
func (t *Trace) Filter(ctx context.Context, filters storage.TraceListFilter) (traces []*storage.Trace, err error) {
	query := t.DB().NewSelect().
		Model(&traces)

	query = traceListFilter(query, filters)

	err = t.DB().NewSelect().
		ColumnExpr("trace.*").
		ColumnExpr("tx.id as tx__id, tx.hash AS tx__hash").
		ColumnExpr("from_addr.id AS from_address__id, from_addr.height AS from_address__height, from_addr.last_height AS from_address__last_height, from_addr.hash AS from_address__hash, from_addr.is_contract AS from_address__is_contract").
		ColumnExpr("to_addr.id AS to_address__id, to_addr.height AS to_address__height, to_addr.last_height AS to_address__last_height, to_addr.hash AS to_address__hash, to_addr.is_contract AS to_address__is_contract").
		ColumnExpr("contract.id AS contract__id, contract_addr.hash AS contract__address__hash, contract.code AS contract__code, contract.verified as contract__verified").
		TableExpr("(?) AS trace", query).
		Join("LEFT JOIN tx ON tx.id = trace.tx_id").
		Join("LEFT JOIN address AS from_addr ON from_addr.id = trace.from_address_id").
		Join("LEFT JOIN address AS to_addr ON to_addr.id = trace.to_address_id").
		Join("LEFT JOIN contract ON contract.id = trace.contract_id").
		Join("LEFT JOIN address AS contract_addr ON contract_addr.id = contract.id").
		Scan(ctx, &traces)

	return
}

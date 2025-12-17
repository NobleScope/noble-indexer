package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type Log struct {
	*postgres.Table[*storage.Log]
}

// NewLog -
func NewLog(db *database.Bun) *Log {
	return &Log{
		Table: postgres.NewTable[*storage.Log](db),
	}
}

// Filter -
func (l *Log) Filter(ctx context.Context, filter storage.LogListFilter) (logs []storage.Log, err error) {
	query := l.DB().NewSelect().
		Model(&logs)

	query = logListFilter(query, filter)
	err = l.DB().NewSelect().
		ColumnExpr("log.*").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("address.address AS address__address").
		TableExpr("(?) AS log", query).
		Join("LEFT JOIN tx ON tx.id = log.tx_id").
		Join("LEFT JOIN address ON address.id = log.address_id").
		Scan(ctx, &logs)

	return
}

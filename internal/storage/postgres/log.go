package postgres

import (
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

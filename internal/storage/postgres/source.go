package postgres

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// Source -
type Source struct {
	*postgres.Table[*storage.Source]
}

// NewSource -
func NewSource(db *database.Bun) *Source {
	return &Source{
		Table: postgres.NewTable[*storage.Source](db),
	}
}

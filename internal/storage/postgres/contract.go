package postgres

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// Contract -
type Contract struct {
	*postgres.Table[*storage.Contract]
}

// NewContract -
func NewContract(db *database.Bun) *Contract {
	return &Contract{
		Table: postgres.NewTable[*storage.Contract](db),
	}
}

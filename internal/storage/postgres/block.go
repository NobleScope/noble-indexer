package postgres

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type Block struct {
	*postgres.Table[*storage.Block]
}

// NewBlock -
func NewBlock(db *database.Bun) *Block {
	return &Block{
		Table: postgres.NewTable[*storage.Block](db),
	}
}

package postgres

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type ERC4337UserOps struct {
	*postgres.Table[*storage.ERC4337UserOp]
}

// NewERC4337UserOps -
func NewERC4337UserOps(db *database.Bun) *ERC4337UserOps {
	return &ERC4337UserOps{
		Table: postgres.NewTable[*storage.ERC4337UserOp](db),
	}
}

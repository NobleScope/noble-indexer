package postgres

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type UserOps struct {
	*postgres.Table[*storage.UserOp]
}

// NewUserOps -
func NewUserOps(db *database.Bun) *UserOps {
	return &UserOps{
		Table: postgres.NewTable[*storage.UserOp](db),
	}
}

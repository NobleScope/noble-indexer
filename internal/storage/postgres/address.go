package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// Address -
type Address struct {
	*postgres.Table[*storage.Address]
}

// NewAddress -
func NewAddress(db *database.Bun) *Address {
	return &Address{
		Table: postgres.NewTable[*storage.Address](db),
	}
}

// ByHash -
func (b *Address) ByHash(ctx context.Context, hash string) (address storage.Address, err error) {
	err = b.DB().NewSelect().Model(&address).
		Where("address = ?", hash).
		Scan(ctx)
	return
}

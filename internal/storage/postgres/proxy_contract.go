package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type ProxyContract struct {
	*postgres.Table[*storage.ProxyContract]
}

// NewProxyContract -
func NewProxyContract(db *database.Bun) *ProxyContract {
	return &ProxyContract{
		Table: postgres.NewTable[*storage.ProxyContract](db),
	}
}

// NotResolved -
func (p *ProxyContract) NotResolved(ctx context.Context) (contracts []storage.ProxyContract, err error) {
	err = p.DB().NewSelect().
		Model(&contracts).
		Relation("Contract.Address").
		Where("proxy_contract.status = ?", types.New).
		Order("height DESC").
		Limit(100).
		Scan(ctx)
	return
}

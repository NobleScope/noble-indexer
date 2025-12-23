package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
	"github.com/uptrace/bun"
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

// FilteredList -
func (p *ProxyContract) FilteredList(
	ctx context.Context,
	filters storage.ListProxyFilters,
) (contracts []storage.ProxyContract, err error) {
	query := p.DB().NewSelect().
		Model(&contracts)

	if filters.Height != 0 {
		query = query.Where("proxy_contract.height = ?", filters.Height)
	}
	if filters.ImplementationId != 0 {
		query = query.Where("proxy_contract.implementation_id = ?", filters.ImplementationId)
	}
	if len(filters.Type) > 0 {
		query = query.Where("proxy_contract.type IN (?)", bun.In(filters.Type))
	}
	if len(filters.Status) > 0 {
		query = query.Where("proxy_contract.status IN (?)", bun.In(filters.Status))
	}

	query = sortScope(query, "height", filters.Sort)
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}
	query = limitScope(query, filters.Limit)

	err = p.DB().NewSelect().TableExpr("(?) AS proxy", query).
		ColumnExpr("proxy.*").
		ColumnExpr("impl_addr.hash AS implementation__address__hash").
		ColumnExpr("contract_addr.hash AS contract__address__hash").
		Join("LEFT JOIN address AS impl_addr ON impl_addr.id = proxy.implementation_id").
		Join("LEFT JOIN address AS contract_addr ON contract_addr.id = proxy.id").
		Scan(ctx, &contracts)

	return
}

package postgres

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

func limitScope(q *bun.SelectQuery, limit int) *bun.SelectQuery {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return q.Limit(limit)
}

func sortScope(q *bun.SelectQuery, field string, sort sdk.SortOrder) *bun.SelectQuery {
	if sort != sdk.SortOrderAsc && sort != sdk.SortOrderDesc {
		sort = sdk.SortOrderAsc
	}
	return q.OrderExpr("? ?", bun.Ident(field), bun.Safe(sort))
}

func addressListFilter(query *bun.SelectQuery, fltrs storage.AddressListFilter) *bun.SelectQuery {
	if fltrs.OnlyContracts {
		query = query.Where("is_contract = ?", true)
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)

	switch fltrs.SortField {
	case "id", "value", "last_height":
		query = sortScope(query, fltrs.SortField, fltrs.Sort)
	case "first_height":
		query = sortScope(query, "height", fltrs.Sort)
	default:
		query = sortScope(query, "id", fltrs.Sort)
	}

	return query
}

func contractListFilter(query *bun.SelectQuery, fltrs storage.ContractListFilter) *bun.SelectQuery {
	if fltrs.TxId != nil {
		query = query.Where("tx_id = ?", fltrs.TxId)
	}

	if fltrs.IsVerified {
		query = query.Where("is_verified = ?", true)
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)

	switch fltrs.SortField {
	case "id", "height":
		query = sortScope(query, fltrs.SortField, fltrs.Sort)
	default:
		query = sortScope(query, "id", fltrs.Sort)
	}

	return query
}

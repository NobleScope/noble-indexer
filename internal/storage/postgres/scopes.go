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

func sortTimeIDScope(q *bun.SelectQuery, sort sdk.SortOrder) *bun.SelectQuery {

	if sort != sdk.SortOrderAsc && sort != sdk.SortOrderDesc {
		sort = sdk.SortOrderAsc
	}

	return q.OrderExpr("time ?0, id ?0", bun.Safe(sort))
}

func addressListFilter(query *bun.SelectQuery, fltrs storage.AddressListFilter) *bun.SelectQuery {
	if fltrs.OnlyContracts {
		query = query.Where("is_contract = ?", true)
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)

	switch fltrs.SortField {
	case "id", "value", "first_height", "last_height":
		query = sortScope(query, fltrs.SortField, fltrs.Sort)
	default:
		query = sortScope(query, "id", fltrs.Sort)
	}

	return query
}

func balanceListFilter(query *bun.SelectQuery, fltrs storage.AddressListFilter) *bun.SelectQuery {
	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)
	query = sortScope(query, "value", fltrs.Sort)
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

func traceListFilter(query *bun.SelectQuery, fltrs storage.TraceListFilter) *bun.SelectQuery {
	if fltrs.TxId != nil {
		query = query.Where("tx_id = ?", *fltrs.TxId)
	}
	if fltrs.AddressFromId != nil {
		query = query.Where("from_address_id = ?", *fltrs.AddressFromId)
	}
	if fltrs.AddressToId != nil {
		query = query.Where("to_address_id = ?", *fltrs.AddressToId)
	}
	if fltrs.ContractId != nil {
		query = query.Where("contract_id = ?", *fltrs.ContractId)
	}
	if fltrs.Height != nil {
		query = query.Where("height = ?", *fltrs.Height)
	}

	if len(fltrs.Type) > 0 {
		query = query.Where("type IN (?)", bun.In(fltrs.Type))
	}
	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)
	query = sortTimeIDScope(query, fltrs.Sort)

	return query
}

func tokenListFilter(query *bun.SelectQuery, fltrs storage.TokenListFilter) *bun.SelectQuery {
	if fltrs.ContractId != nil {
		query = query.Where("contract_id = ?", *fltrs.ContractId)
	}

	if len(fltrs.Type) > 0 {
		query = query.Where("type IN (?)", bun.In(fltrs.Type))
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)
	query = sortScope(query, "id", fltrs.Sort)

	return query
}

func transferListFilter(query *bun.SelectQuery, fltrs storage.TransferListFilter) *bun.SelectQuery {
	if fltrs.TokenId != nil {
		query = query.Where("token_id = ?", *fltrs.TokenId)
	}
	if fltrs.TxId != nil {
		query = query.Where("tx_id = ?", *fltrs.TxId)
	}
	if fltrs.AddressFromId != nil {
		query = query.Where("from_address_id = ?", *fltrs.AddressFromId)
	}
	if fltrs.AddressToId != nil {
		query = query.Where("to_address_id = ?", *fltrs.AddressToId)
	}
	if fltrs.ContractId != nil {
		query = query.Where("contract_id = ?", *fltrs.ContractId)
	}
	if fltrs.Height != nil {
		query = query.Where("height = ?", *fltrs.Height)
	}
	if len(fltrs.Type) > 0 {
		query = query.Where("type IN (?)", bun.In(fltrs.Type))
	}

	if !fltrs.TimeFrom.IsZero() {
		query = query.Where("time >= ?", fltrs.TimeFrom)
	}
	if !fltrs.TimeTo.IsZero() {
		query = query.Where("time < ?", fltrs.TimeTo)
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)
	query = sortTimeIDScope(query, fltrs.Sort)

	return query
}

func tokenBalanceListFilter(query *bun.SelectQuery, fltrs storage.TokenBalanceListFilter) *bun.SelectQuery {
	if fltrs.TokenId != nil {
		query = query.Where("token_id = ?", *fltrs.TokenId)
	}
	if fltrs.AddressId != nil {
		query = query.Where("address_id = ?", *fltrs.AddressId)
	}
	if fltrs.ContractId != nil {
		query = query.Where("contract_id = ?", *fltrs.ContractId)
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)
	query = sortScope(query, "balance", fltrs.Sort)

	return query
}

func logListFilter(query *bun.SelectQuery, fltrs storage.LogListFilter) *bun.SelectQuery {
	if fltrs.TxId != nil {
		query = query.Where("tx_id = ?", *fltrs.TxId)
	}
	if fltrs.AddressId != nil {
		query = query.Where("address_id = ?", *fltrs.AddressId)
	}
	if fltrs.Height != nil {
		query = query.Where("height = ?", *fltrs.Height)
	}

	if !fltrs.TimeFrom.IsZero() {
		query = query.Where("time >= ?", fltrs.TimeFrom)
	}
	if !fltrs.TimeTo.IsZero() {
		query = query.Where("time < ?", fltrs.TimeTo)
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)
	query = sortTimeIDScope(query, fltrs.Sort)

	return query
}

func txListFilter(query *bun.SelectQuery, fltrs storage.TxListFilter) *bun.SelectQuery {
	if fltrs.AddressFromId != nil {
		query = query.Where("from_address_id = ?", *fltrs.AddressFromId)
	}
	if fltrs.AddressToId != nil {
		query = query.Where("to_address_id = ?", *fltrs.AddressToId)
	}
	if fltrs.ContractId != nil {
		query = query.Where("from_address_id = ?", *fltrs.ContractId).
			WhereOr("to_address_id = ?", *fltrs.ContractId)
	}
	if fltrs.Height != nil {
		query = query.Where("height = ?", *fltrs.Height)
	}
	if !fltrs.TimeFrom.IsZero() {
		query = query.Where("time >= ?", fltrs.TimeFrom)
	}
	if !fltrs.TimeTo.IsZero() {
		query = query.Where("time < ?", fltrs.TimeTo)
	}

	if len(fltrs.Type) > 0 {
		query = query.Where("type IN (?)", bun.In(fltrs.Type))
	}
	if len(fltrs.Status) > 0 {
		query = query.Where("status IN (?)", bun.In(fltrs.Status))
	}

	query = limitScope(query, fltrs.Limit)
	query = query.Offset(fltrs.Offset)
	query = sortTimeIDScope(query, fltrs.Sort)

	return query
}

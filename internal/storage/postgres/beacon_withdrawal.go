package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type BeaconWithdrawal struct {
	*postgres.Table[*storage.BeaconWithdrawal]
}

// NewBeaconWithdrawal -
func NewBeaconWithdrawal(db *database.Bun) *BeaconWithdrawal {
	return &BeaconWithdrawal{
		Table: postgres.NewTable[*storage.BeaconWithdrawal](db),
	}
}

// Filter -
func (b *BeaconWithdrawal) Filter(ctx context.Context, filter storage.BeaconWithdrawalListFilter) (beaconWithdrawals []*storage.BeaconWithdrawal, err error) {
	subQuery := b.DB().NewSelect().Model(&beaconWithdrawals)

	if filter.Height != nil {
		subQuery.Where("height = ?", *filter.Height)
	}
	if filter.AddressId != nil {
		subQuery.Where("address_id = ?", *filter.AddressId)
	}

	subQuery = limitScope(subQuery, filter.Limit)
	subQuery = sortTimeIDScope(subQuery, filter.Sort)
	if filter.Offset > 0 {
		subQuery.Offset(filter.Offset)
	}

	query := b.DB().NewSelect().
		TableExpr("(?) AS beacon_withdrawals", subQuery).
		ColumnExpr("beacon_withdrawals.*").
		ColumnExpr("address.hash as address__hash").
		Join("left join address on address.id = beacon_withdrawals.address_id")

	query = sortTimeIDScope(query, filter.Sort)
	err = query.Scan(ctx, &beaconWithdrawals)
	return
}

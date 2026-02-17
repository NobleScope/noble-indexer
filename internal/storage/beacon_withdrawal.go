package storage

import (
	"context"
	"time"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type BeaconWithdrawalListFilter struct {
	Limit     int
	Offset    int
	Sort      storage.SortOrder
	Height    *pkgTypes.Level
	AddressId *uint64
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IBeaconWithdrawal interface {
	storage.Table[*BeaconWithdrawal]

	Filter(ctx context.Context, filter BeaconWithdrawalListFilter) (beaconWithdrawals []*BeaconWithdrawal, err error)
}

// BeaconWithdrawal -
type BeaconWithdrawal struct {
	bun.BaseModel `bun:"beacon_withdrawal" comment:"Table with beacon chain withdrawals."`

	Id             uint64          `bun:",pk,notnull,autoincrement" comment:"Unique internal identity"`
	Height         pkgTypes.Level  `bun:"height"                    comment:"The number (height) of block"`
	Time           time.Time       `bun:"time,pk,notnull"           comment:"The time of block"`
	Index          int64           `bun:"index"                     comment:"The index of the withdrawal in the block"`
	ValidatorIndex int64           `bun:"validator_index"           comment:"The index of the validator making the withdrawal"`
	AddressId      uint64          `bun:"address_id"                comment:"The address to which the withdrawal is sent"`
	Amount         decimal.Decimal `bun:"amount,type:numeric"       comment:"The amount of withdrawn"`

	Address Address `bun:"rel:belongs-to,join:address_id=id"`
}

func (bw *BeaconWithdrawal) TableName() string {
	return "beacon_withdrawal"
}

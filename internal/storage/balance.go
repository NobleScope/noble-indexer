package storage

import (
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IBalance interface {
	storage.Table[*Balance]
}

type Balance struct {
	bun.BaseModel `bun:"balance" comment:"Table with account balances."`

	Id    uint64          `bun:"id,pk,notnull"      comment:"Unique internal identity"`
	Value decimal.Decimal `bun:"value,type:numeric" comment:"Balance value"`
}

func (Balance) TableName() string {
	return "balance"
}

func EmptyBalance() *Balance {
	return &Balance{
		Value: decimal.Zero,
	}
}

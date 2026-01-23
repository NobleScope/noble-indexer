package storage

import (
	"context"

	"fmt"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type TokenBalanceListFilter struct {
	Limit      int
	Offset     int
	Sort       storage.SortOrder
	AddressId  *uint64
	ContractId *uint64
	TokenId    *decimal.Decimal
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ITokenBalance interface {
	storage.Table[*TokenBalance]

	Filter(ctx context.Context, filter TokenBalanceListFilter) ([]TokenBalance, error)
}

// TokenBalance -
type TokenBalance struct {
	bun.BaseModel `bun:"token_balance" comment:"Table with addresses token balances"`

	Id         uint64          `bun:",pk,autoincrement"                              comment:"Unique internal identity"`
	TokenID    decimal.Decimal `bun:"token_id,type:numeric,unique:token_balance_idx" comment:"Token ID"`
	ContractID uint64          `bun:"contract_id,unique:token_balance_idx"           comment:"Contract address id"`
	AddressID  uint64          `bun:"address_id,unique:token_balance_idx"            comment:"Address ID"`
	Balance    decimal.Decimal `bun:"balance,type:numeric"                           comment:"Token balance"`

	Token    Token    `bun:"rel:belongs-to,join:token_id=token_id,join:contract_id=contract_id"`
	Contract Contract `bun:"rel:belongs-to,join:contract_id=id"`
	Address  Address  `bun:"rel:belongs-to,join:address_id=id"`
}

// TableName -
func (TokenBalance) TableName() string {
	return "token_balance"
}

func (t TokenBalance) String() string {
	return fmt.Sprintf("%s:%s:%s", t.Contract.String(), t.TokenID.String(), t.Address.String())
}

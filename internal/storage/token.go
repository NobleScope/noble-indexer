package storage

import (
	"fmt"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IToken interface {
	storage.Table[*Token]
}

// Token -
type Token struct {
	bun.BaseModel `bun:"token" comment:"Table with tokens"`

	TokenID        decimal.Decimal `bun:"token_id,pk,type:numeric" comment:"Token ID"`
	ContractId     uint64          `bun:"contract_id,pk"           comment:"Contract address id"`
	Type           types.TokenType `bun:",type:token_type"         comment:"Token type"`
	Name           string          `bun:"name"                     comment:"Token name"`
	Symbol         string          `bun:"symbol"                   comment:"Token symbol"`
	Decimals       uint            `bun:"decimals"                 comment:"Decimals"`
	TransfersCount uint64          `bun:"transfers_count"          comment:"Transfers count"`
	Supply         decimal.Decimal `bun:"supply,type:numeric"      comment:"Token supply"`

	Contract Contract `bun:"rel:belongs-to,join:contract_id=id"`
}

// TableName -
func (Token) TableName() string {
	return "token"
}

func (t Token) String() string {
	return fmt.Sprintf("%s:%s", t.Contract.String(), t.TokenID.String())
}

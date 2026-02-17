package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type TokenListFilter struct {
	Limit      int
	Offset     int
	Sort       storage.SortOrder
	ContractId *uint64
	Type       []types.TokenType
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IToken interface {
	storage.Table[*Token]

	Get(ctx context.Context, contractId uint64, tokenId decimal.Decimal) (Token, error)
	Filter(ctx context.Context, filter TokenListFilter) ([]Token, error)
	PendingMetadata(ctx context.Context, delay time.Duration, limit int) ([]*Token, error)
}

// Token -
type Token struct {
	bun.BaseModel `bun:"token" comment:"Table with tokens"`

	Id             uint64               `bun:"id,notnull,autoincrement"         comment:"Unique internal identity"`
	TokenID        decimal.Decimal      `bun:"token_id,pk,type:numeric"         comment:"Token ID"`
	ContractId     uint64               `bun:"contract_id,pk"                   comment:"Contract address id"`
	Type           types.TokenType      `bun:",type:token_type"                 comment:"Token type"`
	Height         pkgTypes.Level       `bun:"height"                           comment:"Block number of the first token occurrence"`
	LastHeight     pkgTypes.Level       `bun:"last_height"                      comment:"Block number of the last token occurrence"`
	Name           string               `bun:"name"                             comment:"Token name"`
	Symbol         string               `bun:"symbol"                           comment:"Token symbol"`
	Decimals       uint8                `bun:"decimals"                         comment:"Decimals"`
	TransfersCount uint64               `bun:"transfers_count"                  comment:"Transfers count"`
	Supply         decimal.Decimal      `bun:"supply,type:numeric"              comment:"Token supply"`
	MetadataLink   string               `bun:"metadata_link"                    comment:"Metadata link"`
	Status         types.MetadataStatus `bun:",type:metadata_status"            comment:"Token metadata status"`
	RetryCount     uint64               `bun:"retry_count"                      comment:"Retry count to resolve metadata"`
	Error          string               `bun:"error"                            comment:"Error"`
	Metadata       []byte               `bun:"metadata,type:bytea"              comment:"Token metadata"`
	UpdatedAt      time.Time            `bun:"updated_at,notnull,default:now()" comment:"last update time"`
	Logo           string               `bun:"logo"                             comment:"Logo URL"`

	Contract Contract `bun:"rel:belongs-to,join:contract_id=id"`
}

// TableName -
func (Token) TableName() string {
	return "token"
}

func (t Token) String() string {
	return fmt.Sprintf("%s:%s", t.Contract.String(), t.TokenID.String())
}

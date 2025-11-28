package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IToken interface {
	storage.Table[*Token]

	PendingMetadata(ctx context.Context, delay time.Duration, limit int) ([]*Token, error)
}

// Token -
type Token struct {
	bun.BaseModel `bun:"token" comment:"Table with tokens"`

	Id             uint64               `bun:"id,notnull,autoincrement"         comment:"Unique internal identity"`
	TokenID        decimal.Decimal      `bun:"token_id,pk,type:numeric"         comment:"Token ID"`
	ContractId     uint64               `bun:"contract_id,pk"                   comment:"Contract address id"`
	Type           types.TokenType      `bun:",type:token_type"                 comment:"Token type"`
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

	Contract Contract `bun:"rel:belongs-to,join:contract_id=id"`
}

// TableName -
func (Token) TableName() string {
	return "token"
}

func (t Token) String() string {
	return fmt.Sprintf("%s:%s", t.Contract.String(), t.TokenID.String())
}

package storage

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ITx interface {
	storage.Table[*Tx]

	ByHeight(ctx context.Context, height pkgTypes.Level, limit, offset int, order storage.SortOrder) (txs []*Tx, err error)
	ByHash(ctx context.Context, hash pkgTypes.Hex) (tx Tx, err error)
}

// Tx -
type Tx struct {
	bun.BaseModel `bun:"tx" comment:"Table with transactions."`

	Id     uint64         `bun:"id,autoincrement,pk,notnull" comment:"Unique internal id"`
	Height pkgTypes.Level `bun:",notnull"                    comment:"The number (height) of this block"`
	Time   time.Time      `bun:"time,pk,notnull"             comment:"The time of block"`

	Gas      decimal.Decimal `bun:"gas,type:numeric"       comment:"Gas"`
	GasPrice decimal.Decimal `bun:"gas_price,type:numeric" comment:"Gas price"`
	Hash     pkgTypes.Hex    `bun:"hash"                   comment:"Transaction hash"`
	Nonce    int64           `bun:"nonce"                  comment:"Nonce"`
	Index    int64           `bun:"index"                  comment:"Transaction index in the block"`
	Amount   decimal.Decimal `bun:"amount,type:numeric"    comment:"Value in Wei"`
	Type     types.TxType    `bun:",type:tx_type"          comment:"Transaction type"`
	Input    []byte          `bun:"input"                  comment:"Transaction input"`

	ContractId        *uint64         `bun:"contract_id"                      comment:"Contract address id"`
	CumulativeGasUsed decimal.Decimal `bun:"cumulative_gas_used,type:numeric" comment:"Cumulative gas used"`
	EffectiveGasPrice decimal.Decimal `bun:"effective_gas_price,type:numeric" comment:"Effective gas price"`
	FromAddressId     uint64          `bun:"from_address_id"                  comment:"From address id"`
	ToAddressId       *uint64         `bun:"to_address_id"                    comment:"To address id"`
	Fee               decimal.Decimal `bun:"fee,type:numeric"                 comment:"Fee in Wei"`
	GasUsed           decimal.Decimal `bun:"gas_used,type:numeric"            comment:"Gas used"`
	Status            types.TxStatus  `bun:"status,type:tx_status"            comment:"Transaction status"`
	LogsBloom         []byte          `bun:"logs_bloom"                       comment:"Logs bloom"`

	LogsCount   int `bun:"logs_count"   comment:"Logs count"`
	TracesCount int `bun:"traces_count" comment:"Traces count"`

	Contract    *Contract `bun:"rel:belongs-to,join:contract_id=id"`
	FromAddress Address   `bun:"rel:belongs-to,join:from_address_id=id"`
	ToAddress   *Address  `bun:"rel:belongs-to,join:to_address_id=id"`
	Logs        []Log     `bun:"rel:has-many"`
}

// TableName -
func (Tx) TableName() string {
	return "tx"
}

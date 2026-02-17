package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type TxListFilter struct {
	Limit         int
	Offset        int
	Sort          storage.SortOrder
	Height        *uint64
	Type          []types.TxType
	Status        []types.TxStatus
	AddressFromId *uint64
	AddressToId   *uint64
	AddressId     *uint64
	ContractId    *uint64
	TimeFrom      time.Time
	TimeTo        time.Time
	WithABI       bool
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ITx interface {
	storage.Table[*Tx]

	ByHeight(ctx context.Context, height pkgTypes.Level, limit, offset int, order storage.SortOrder) ([]*Tx, error)
	ByHash(ctx context.Context, hash pkgTypes.Hex, withABI bool) (Tx, error)
	Filter(ctx context.Context, filter TxListFilter) ([]Tx, error)
}

// Tx -
type Tx struct {
	bun.BaseModel `bun:"tx" comment:"Table with transactions."`

	Id     uint64         `bun:"id,autoincrement,pk,notnull" comment:"Unique internal id"`
	Height pkgTypes.Level `bun:",notnull"                    comment:"The number (height) of this block"`
	Time   time.Time      `bun:"time,pk,notnull"             comment:"The time of block"`

	Gas      decimal.Decimal `bun:"gas,type:numeric"       comment:"Gas"`
	GasPrice decimal.Decimal `bun:"gas_price,type:numeric" comment:"Gas price"`
	Hash     pkgTypes.Hex    `bun:"hash,type:bytea"        comment:"Transaction hash"`
	Nonce    int64           `bun:"nonce"                  comment:"Nonce"`
	Index    int64           `bun:"index"                  comment:"Transaction index in the block"`
	Amount   decimal.Decimal `bun:"amount,type:numeric"    comment:"Value in Wei"`
	Type     types.TxType    `bun:",type:tx_type"          comment:"Transaction type"`
	Input    []byte          `bun:"input"                  comment:"Transaction input"`

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

	FromAddress Address     `bun:"rel:belongs-to,join:from_address_id=id"`
	ToAddress   *Address    `bun:"rel:belongs-to,join:to_address_id=id"`
	Logs        []*Log      `bun:"rel:has-many"`
	Transfers   []*Transfer `bun:"rel:has-many"`

	ToContractABI json.RawMessage `bun:"to_contract_abi,scanonly"`
}

// TableName -
func (Tx) TableName() string {
	return "tx"
}

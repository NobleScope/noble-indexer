package storage

import (
	"context"
	"time"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

type LogListFilter struct {
	Limit     int
	Offset    int
	Sort      storage.SortOrder
	Height    *uint64
	TxId      *uint64
	AddressId *uint64
	TimeFrom  time.Time
	TimeTo    time.Time
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ILog interface {
	storage.Table[*Log]

	Filter(ctx context.Context, filter LogListFilter) ([]Log, error)
}

// Log -
type Log struct {
	bun.BaseModel `bun:"log" comment:"Table with logs."`

	Id        uint64         `bun:"id,pk,notnull,autoincrement" comment:"Unique internal id"`
	Height    pkgTypes.Level `bun:"height,notnull"              comment:"The number (height) of this block"`
	Time      time.Time      `bun:"time,pk,notnull"             comment:"The time of block"`
	Index     int64          `bun:"index"                       comment:"Index in transaction"`
	Name      string         `bun:"name"                        comment:"Log name"`
	TxId      uint64         `bun:"tx_id"                       comment:"Transaction id"`
	Data      pkgTypes.Hex   `bun:"data"                        comment:"Log data"`
	Topics    []pkgTypes.Hex `bun:"topics,type:bytea"           comment:"Log topics"`
	AddressId uint64         `bun:"address_id"                  comment:"Contract address ID whose invocation generated this log"`
	Removed   bool           `bun:"removed"                     comment:"Removed during the reorg"`

	Address Address `bun:"rel:belongs-to,join:address_id=id"`
	Tx      Tx      `bun:"rel:belongs-to,join:tx_id=id"`
}

// TableName -
func (Log) TableName() string {
	return "log"
}

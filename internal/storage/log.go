package storage

import (
	"time"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ILog interface {
	storage.Table[*Log]
}

// Log -
type Log struct {
	bun.BaseModel `bun:"log" comment:"Table with logs."`

	Id      uint64         `bun:"id,pk,notnull,autoincrement" comment:"Unique internal id"`
	Height  pkgTypes.Level `bun:"height,notnull"              comment:"The number (height) of this block"`
	Time    time.Time      `bun:"time,pk,notnull"             comment:"The time of block"`
	Index   int64          `bun:"index"                       comment:"Index in transaction"`
	Name    string         `bun:"name"                        comment:"Log name"`
	TxId    uint64         `bun:"tx_id"                       comment:"Transaction id"`
	Data    []byte         `bun:"data"                        comment:"Log data"`
	Topics  []pkgTypes.Hex `bun:"topics,type:bytea"           comment:"Log topics"`
	Address pkgTypes.Hex   `bun:"address,type:bytea"          comment:"Contract address whose invocation generated this log"`

	Removed bool `bun:"removed" comment:"Removed during the reorg"`
}

// TableName -
func (Log) TableName() string {
	return "log"
}

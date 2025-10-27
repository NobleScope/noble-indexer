package storage

import (
	"context"
	"time"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IBlockStats interface {
	ByHeight(ctx context.Context, height pkgTypes.Level) (BlockStats, error)
}

type BlockStats struct {
	bun.BaseModel `bun:"table:block_stats" comment:"Table with block stats."`

	Id        uint64         `bun:",pk,notnull,autoincrement" comment:"Unique internal identity"`
	Height    pkgTypes.Level `bun:"height"                    comment:"The number (height) of this block"`
	Time      time.Time      `bun:"time,pk,notnull"           comment:"The time of block"`
	TxCount   int64          `bun:"tx_count"                  comment:"Count of transactions in block"`
	BlockTime uint64         `bun:"block_time"                comment:"Time in milliseconds between current and previous block"`
}

func (BlockStats) TableName() string {
	return "block_stats"
}

package storage

import (
	"context"
	"time"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type BlockListFilter struct {
	Limit     int
	Offset    int
	Sort      storage.SortOrder
	WithStats bool
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IBlock interface {
	storage.Table[*Block]

	Last(ctx context.Context) (Block, error)
	ByHeight(ctx context.Context, height pkgTypes.Level, withStats bool) (Block, error)
	Filter(ctx context.Context, filters BlockListFilter) ([]Block, error)
}

// Block -
type Block struct {
	bun.BaseModel `bun:"block" comment:"Table with blocks."`

	Id            uint64          `bun:",pk,notnull,autoincrement"     comment:"Unique internal identity"`
	Height        pkgTypes.Level  `bun:"height"                        comment:"The number (height) of this block"`
	Time          time.Time       `bun:"time,pk,notnull"               comment:"The time of block"`
	GasLimit      decimal.Decimal `bun:"gas_limit,type:numeric"        comment:"Gas limit"`
	GasUsed       decimal.Decimal `bun:"gas_used,type:numeric"         comment:"Gas used"`
	BaseFeePerGas uint64          `bun:"base_fee_per_gas,type:numeric" comment:"Fee per gas"`
	MinerId       uint64          `bun:"miner_id"                      comment:"Miner address id"`

	DifficultyHash       pkgTypes.Hex `bun:"difficulty_hash,type:bytea"        comment:"Difficulty hash"`
	ExtraDataHash        pkgTypes.Hex `bun:"extra_data_hash,type:bytea"        comment:"Extra data hash"`
	Hash                 pkgTypes.Hex `bun:"hash,type:bytea"                   comment:"Block hash"`
	LogsBloomHash        pkgTypes.Hex `bun:"logs_bloom_hash,type:bytea"        comment:"Logs bloom hash"`
	MixHash              pkgTypes.Hex `bun:"mix_hash,type:bytea"               comment:"Mix hash"`
	NonceHash            pkgTypes.Hex `bun:"nonce_hash,type:bytea"             comment:"Nonce hash"`
	ParentHashHash       pkgTypes.Hex `bun:"parent_hash,type:bytea"            comment:"Hash of parent block"`
	ReceiptsRootHash     pkgTypes.Hex `bun:"receipts_root_hash,type:bytea"     comment:"Hash of receipts root"`
	Sha3UnclesHash       pkgTypes.Hex `bun:"sha3_uncles_hash,type:bytea"       comment:"Sha3 hash"`
	SizeHash             pkgTypes.Hex `bun:"size_hash,type:bytea"              comment:"Size of block in bytes"`
	StateRootHash        pkgTypes.Hex `bun:"state_root_hash,type:bytea"        comment:"Hash of state root"`
	TransactionsRootHash pkgTypes.Hex `bun:"transactions_root_hash,type:bytea" comment:"Hash of transactions root"`

	Txs         []*Tx               `bun:"rel:has-many"                    json:"-"`
	Traces      []*Trace            `bun:"rel:has-many"                    json:"-"`
	Withdrawals []*BeaconWithdrawal `bun:"rel:has-many"                    json:"-"`
	Miner       Address             `bun:"rel:belongs-to,join:miner_id=id"`
	Stats       *BlockStats         `bun:"rel:has-one,join:height=height"`
}

// TableName -
func (Block) TableName() string {
	return "block"
}

package storage

import (
	"time"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IBlock interface {
	storage.Table[*Block]
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

	DifficultyHash       pkgTypes.Hex `bun:"difficulty_hash"        comment:"Difficulty hash"`
	ExtraDataHash        pkgTypes.Hex `bun:"extra_data_hash"        comment:"Extra data hash"`
	Hash                 pkgTypes.Hex `bun:"hash"                   comment:"Block hash"`
	LogsBloomHash        pkgTypes.Hex `bun:"logs_bloom_hash"        comment:"Logs bloom hash"`
	MinerHash            pkgTypes.Hex `bun:"miner_hash"             comment:"Miner address hash"`
	MixHash              pkgTypes.Hex `bun:"mix_hash"               comment:"Mix hash"`
	NonceHash            pkgTypes.Hex `bun:"nonce_hash"             comment:"Nonce hash"`
	ParentHashHash       pkgTypes.Hex `bun:"parent_hash"            comment:"Hash of parent block"`
	ReceiptsRootHash     pkgTypes.Hex `bun:"receipts_root_hash"     comment:"Hash of receipts root"`
	Sha3UnclesHash       pkgTypes.Hex `bun:"sha3_uncles_hash"       comment:"Sha3 hash"`
	SizeHash             pkgTypes.Hex `bun:"size_hash"              comment:"Size of block in bytes"`
	StateRootHash        pkgTypes.Hex `bun:"state_root_hash"        comment:"Hash of state root"`
	TotalDifficultyHash  pkgTypes.Hex `bun:"total_difficulty_hash"  comment:"Total difficulty hash"`
	TransactionsRootHash pkgTypes.Hex `bun:"transactions_root_hash" comment:"Hash of transactions root"`

	Txs    []*Tx    `bun:"rel:has-many" json:"-"`
	Traces []*Trace `bun:"rel:has-many" json:"-"`
}

// TableName -
func (Block) TableName() string {
	return "block"
}

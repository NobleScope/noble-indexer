package storage

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ITrace interface {
	storage.Table[*Trace]
}

// Trace -
type Trace struct {
	bun.BaseModel `bun:"trace" comment:"Table with tx traces."`

	Id     uint64         `bun:",pk,notnull,autoincrement" comment:"Unique internal identity"`
	Height pkgTypes.Level `bun:"height"                    comment:"The number (height) of block"`
	Time   time.Time      `bun:"time,pk,notnull"           comment:"The time of block"`

	TxId uint64  `bun:"tx_id"           comment:"Transaction identity"`
	From uint64  `bun:"from_address_id" comment:"From address identity"`
	To   *uint64 `bun:"to_address_id"   comment:"To address identity"`

	// init
	GasLimit       uint64          `bun:"gas_limit"       comment:"Gas limit"`
	Value          *uint64         `bun:"value"           comment:"Value in Wei"`
	Input          []byte          `bun:"input"           comment:"Input data"`
	Type           types.TraceType `bun:"type"            comment:"Trace type"`
	InitHash       *pkgTypes.Hex   `bun:"init_hash"       comment:"Code of the contract being created"`
	CreationMethod *string         `bun:"creation_method" comment:"Creation method"`

	// result
	GasUsed    uint64        `bun:"gas_used"    comment:"Gas used"`
	Output     []byte        `bun:"output"      comment:"Output data"`
	ContractId *uint64       `bun:"contract_id" comment:"Address identity of the new contract"`
	Code       *pkgTypes.Hex `bun:"code"        comment:"New contract code"`

	Subtraces uint64 `bun:"subtraces" comment:"Amount of subtraces"`

	FromAddress Address  `bun:"-"`
	ToAddress   *Address `bun:"-"`
	Tx          Tx       `bun:"-"`
}

// TableName -
func (Trace) TableName() string {
	return "trace"
}

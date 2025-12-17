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

type TraceListFilter struct {
	Limit         int
	Offset        int
	Sort          storage.SortOrder
	Height        *uint64
	TxId          *uint64
	AddressFromId *uint64
	AddressToId   *uint64
	ContractId    *uint64
	Type          []types.TraceType
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ITrace interface {
	storage.Table[*Trace]

	Filter(ctx context.Context, filter TraceListFilter) (traces []*Trace, err error)
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

	GasLimit       decimal.Decimal  `bun:"gas_limit,type:numeric" comment:"Gas limit"`
	Amount         *decimal.Decimal `bun:"amount,type:numeric"    comment:"Value in Wei"`
	Input          []byte           `bun:"input"                  comment:"Input data"`
	TxPosition     uint64           `bun:"tx_position"            comment:"Transaction position"`
	TraceAddress   []uint64         `bun:"trace_address"          comment:"Trace position in the call tree"`
	Type           types.TraceType  `bun:"type"                   comment:"Trace type"`
	InitHash       *pkgTypes.Hex    `bun:"init_hash,type:bytea"   comment:"Code of the contract being created"`
	CreationMethod *string          `bun:"creation_method"        comment:"Creation method"`

	GasUsed    decimal.Decimal `bun:"gas_used,type:numeric" comment:"Gas used"`
	Output     []byte          `bun:"output"                comment:"Output data"`
	ContractId *uint64         `bun:"contract_id"           comment:"Address identity of the new contract"`

	Subtraces uint64 `bun:"subtraces" comment:"Amount of subtraces"`

	FromAddress Address   `bun:"rel:belongs-to,join:from_address_id=id"`
	ToAddress   *Address  `bun:"rel:belongs-to,join:to_address_id=id"`
	Contract    *Contract `bun:"rel:belongs-to,join:contract_id=id"`
	Tx          Tx        `bun:"rel:belongs-to,join:tx_id=id"`
}

// TableName -
func (Trace) TableName() string {
	return "trace"
}

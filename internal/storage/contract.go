package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

type ContractListFilter struct {
	Limit      int
	Offset     int
	Sort       storage.SortOrder
	SortField  string
	IsVerified bool
	TxId       *uint64
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IContract interface {
	storage.Table[*Contract]

	ByTxId(ctx context.Context, id uint64) (Contract, error)
	ListWithTx(ctx context.Context, filters ContractListFilter) ([]Contract, error)
	PendingMetadata(ctx context.Context, delay time.Duration, limit int) ([]*Contract, error)
}

// Contract -
type Contract struct {
	bun.BaseModel `bun:"contract" comment:"Table with contracts."`

	Id               uint64               `bun:"id,pk,notnull"                    comment:"Unique internal identity"`
	Height           pkgTypes.Level       `bun:"height"                           comment:"Block number in which the contract was deployed"`
	Code             pkgTypes.Hex         `bun:"code,type:bytea"                  comment:"Contract code"`
	Verified         bool                 `bun:"verified,default:false,notnull"   comment:"Verified or not"`
	TxId             *uint64              `bun:"tx_id"                            comment:"Transaction in which this contract was deployed"`
	ABI              json.RawMessage      `bun:"abi,type:jsonb,nullzero"          comment:"Contract ABI"`
	CompilerVersion  string               `bun:"compiler_version,notnull"         comment:"Compiler version"`
	MetadataLink     string               `bun:"metadata_link"                    comment:"Metadata link"`
	Language         string               `bun:"language"                         comment:"Language"`
	OptimizerEnabled bool                 `bun:"optimizer_enabled"                comment:"Optimizer enabled"`
	Tags             []string             `bun:"tags,array"                       comment:"Implemented interfaces tags"`
	Status           types.MetadataStatus `bun:",type:metadata_status,nullzero"   comment:"Contract metadata status"`
	RetryCount       uint64               `bun:"retry_count"                      comment:"Retry count to resolve metadata"`
	Error            string               `bun:"error"                            comment:"Error"`
	UpdatedAt        time.Time            `bun:"updated_at,notnull,default:now()" comment:"last update time"`

	Address Address `bun:"rel:belongs-to,join:id=id"`
	Tx      *Tx     `bun:"-"`
}

// TableName -
func (Contract) TableName() string {
	return "contract"
}

func (contract Contract) String() string {
	return contract.Address.Address
}

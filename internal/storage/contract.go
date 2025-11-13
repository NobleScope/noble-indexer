package storage

import (
	"context"
	"encoding/json"

	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IContract interface {
	storage.Table[*Contract]

	ListWithMetadata(ctx context.Context, startId uint64) (contracts []*Contract, err error)
}

// Contract -
type Contract struct {
	bun.BaseModel `bun:"contract" comment:"Table with contracts."`

	Id               uint64          `bun:"id,pk,notnull"                  comment:"Unique internal identity"`
	Address          string          `bun:"address,unique:contract_idx"    comment:"Human-readable address"`
	Code             []byte          `bun:"code"                           comment:"Contract code"`
	Verified         bool            `bun:"verified,default:false,notnull" comment:"Verified or not"`
	TxId             *uint64         `bun:"tx_id"                          comment:"Transaction in which this contract was deployed"`
	ABI              json.RawMessage `bun:"abi,type:jsonb,nullzero"        comment:"Contract ABI"`
	CompilerVersion  string          `bun:"compiler_version,notnull"       comment:"Compiler version"`
	MetadataLink     string          `bun:"metadata_link"                  comment:"Metadata link"`
	Language         string          `bun:"language"                       comment:"Language"`
	OptimizerEnabled bool            `bun:"optimizer_enabled"              comment:"Optimizer enabled"`
	Tags             []string        `bun:"tags,array"                     comment:"Implemented interfaces tags"`

	Tx *Tx `bun:"-"`
}

// TableName -
func (Contract) TableName() string {
	return "contract"
}

func (contract Contract) String() string {
	return contract.Address
}

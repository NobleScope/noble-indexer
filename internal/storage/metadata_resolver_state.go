package storage

import (
	"context"

	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IMetadataResolverState interface {
	storage.Table[*MetadataResolverState]

	ByName(ctx context.Context, name string) (MetadataResolverState, error)
	IsNoRows(err error) bool
}

// MetadataResolverState -
type MetadataResolverState struct {
	bun.BaseModel `bun:"metadata_resolver_state" comment:"Contract metadata resolver state"`

	Id             uint64 `bun:",pk,autoincrement"     comment:"Unique internal identity"`
	Name           string `bun:",unique:resolver_name" comment:"Resolver name"`
	LastContractId uint64 `bun:"last_contract_id"      comment:"Last resolved contract id"`
}

// TableName -
func (MetadataResolverState) TableName() string {
	return "metadata_resolver_state"
}

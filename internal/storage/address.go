package storage

import (
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IAddress interface {
	storage.Table[*Address]
}

// Address -
type Address struct {
	bun.BaseModel `bun:"address" comment:"Table with addresses."`

	Id         uint64      `bun:"id,pk,notnull,autoincrement"       comment:"Unique internal identity"`
	Height     types.Level `bun:"height"                            comment:"Block number of the first address occurrence."`
	LastHeight types.Level `bun:"last_height"                       comment:"Block number of the last address occurrence."`
	Address    string      `bun:"address,unique:address_idx"        comment:"Human-readable address."`
	IsContract bool        `bun:"is_contract,default:false,notnull" comment:"Address is a contract or not."`

	Balance []*Balance `bun:"rel:has-many,join:id=id"`
}

// TableName -
func (Address) TableName() string {
	return "address"
}

func (address Address) String() string {
	return address.Address
}

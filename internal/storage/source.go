package storage

import (
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ISource interface {
	storage.Table[*Source]
}

// Source -
type Source struct {
	bun.BaseModel `bun:"source" comment:"Table with contract sources."`

	Id         uint64   `bun:"id,pk,notnull,autoincrement" comment:"Unique internal identity"`
	Name       string   `bun:"name"                        comment:"Source name"`
	License    string   `bun:"license"                     comment:"License"`
	Urls       []string `bun:"urls,array"                  comment:"Links to sources"`
	Content    string   `bun:"content,type:text"           comment:"Content"`
	ContractId uint64   `bun:"contract_id"                 comment:"Contract id"`
}

// TableName -
func (Source) TableName() string {
	return "source"
}

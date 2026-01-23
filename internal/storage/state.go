package storage

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IState interface {
	storage.Table[*State]

	ByName(ctx context.Context, name string) (State, error)
}

// State -
type State struct {
	bun.BaseModel `bun:"state" comment:"Current indexer state"`

	Id                     uint64      `bun:",pk,autoincrement"        comment:"Unique internal identity"`
	Name                   string      `bun:",unique:state_name"       comment:"Indexer name"`
	LastHeight             types.Level `bun:"last_height"              comment:"Last block height"`
	LastHash               []byte      `bun:"last_hash"                comment:"Last block hash"`
	LastTime               time.Time   `bun:"last_time"                comment:"Time of last block"`
	TotalTx                int64       `bun:"total_tx"                 comment:"Transactions count"`
	TotalAccounts          int64       `bun:"total_accounts"           comment:"Accounts count"`
	TotalContracts         int64       `bun:"total_contracts"          comment:"Contracts count"`
	TotalVerifiedContracts int64       `bun:"total_verified_contracts" comment:"Verified contracts count"`
	TotalTokens            int64       `bun:"total_tokens"             comment:"Tokens count"`
	ChainId                int64       `bun:"chain_id"                 comment:"Noble chain id"`
}

// TableName -
func (State) TableName() string {
	return "state"
}

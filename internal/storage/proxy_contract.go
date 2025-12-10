package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IProxyContract interface {
	storage.Table[*ProxyContract]

	NotResolved(ctx context.Context) (contracts []ProxyContract, err error)
}

// ProxyContract -
type ProxyContract struct {
	bun.BaseModel `bun:"proxy_contract" comment:"Table with proxy contracts"`

	Id                uint64            `bun:"id,pk,notnull"      comment:"Unique internal identity, match contract ID"`
	Height            pkgTypes.Level    `bun:"height"             comment:"Block height"`
	Type              types.ProxyType   `bun:",type:proxy_type"   comment:"Proxy contract type"`
	Status            types.ProxyStatus `bun:",type:proxy_status" comment:"Status of resolved implementation"`
	ResolvingAttempts uint              `bun:"resolving_attempts" comment:"Count of resolving attempts"`
	ImplementationID  *uint64           `bun:"implementation_id"  comment:"Internal implementation contract ID"`

	Contract       Contract  `bun:"rel:belongs-to,join:id=id"`
	Implementation *Contract `bun:"rel:belongs-to,join:implementation_id=id"`
}

// TableName -
func (ProxyContract) TableName() string {
	return "proxy_contract"
}

func (t ProxyContract) String() string {
	return t.Contract.String()
}

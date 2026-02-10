package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type TransferListFilter struct {
	Limit         int
	Offset        int
	Sort          storage.SortOrder
	Height        *uint64
	TxId          *uint64
	Type          []types.TransferType
	AddressFromId *uint64
	AddressToId   *uint64
	ContractId    *uint64
	TokenId       *decimal.Decimal
	TimeFrom      time.Time
	TimeTo        time.Time
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ITransfer interface {
	storage.Table[*Transfer]

	Get(ctx context.Context, id uint64) (Transfer, error)
	Filter(ctx context.Context, filter TransferListFilter) ([]Transfer, error)
}

// Transfer -
type Transfer struct {
	bun.BaseModel `bun:"transfer" comment:"Table with token transfers"`

	Id            uint64             `bun:",pk,autoincrement"     comment:"Unique internal identity"`
	Height        pkgTypes.Level     `bun:"height"                comment:"Block height"`
	Time          time.Time          `bun:"time,pk,notnull"       comment:"Time of block"`
	TokenID       decimal.Decimal    `bun:"token_id,type:numeric" comment:"Token ID"`
	Amount        decimal.Decimal    `bun:"amount,type:numeric"   comment:"Transfer amount"`
	Type          types.TransferType `bun:",type:transfer_type"   comment:"Transfer type"`
	ContractId    uint64             `bun:"contract_id"           comment:"Contract address id"`
	FromAddressId *uint64            `bun:"from_address_id"       comment:"From address id"`
	ToAddressId   *uint64            `bun:"to_address_id"         comment:"To address id"`
	TxID          uint64             `bun:"tx_id"                 comment:"Transaction id"`

	Contract    Contract `bun:"rel:belongs-to,join:contract_id=id"`
	FromAddress *Address `bun:"rel:belongs-to,join:from_address_id=id"`
	ToAddress   *Address `bun:"rel:belongs-to,join:to_address_id=id"`
	Tx          Tx       `bun:"rel:belongs-to,join:tx_id=id"`
	Token       *Token   `bun:"rel:belongs-to,join:token_id=token_id,join:contract_id=contract_id"`
}

// TableName -
func (Transfer) TableName() string {
	return "transfer"
}

func (t Transfer) String() string {
	return fmt.Sprintf("%s:%s", t.Contract.String(), t.TokenID.String())
}

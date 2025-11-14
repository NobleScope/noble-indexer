package storage

import (
	"fmt"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ITransfer interface {
	storage.Table[*Transfer]
}

// Transfer -
type Transfer struct {
	bun.BaseModel `bun:"transfer" comment:"Table with token transfers"`

	Id            uint64             `bun:",pk,autoincrement"   comment:"Unique internal identity"`
	Height        pkgTypes.Level     `bun:"height"              comment:"Block height"`
	Time          time.Time          `bun:"time,pk,notnull"     comment:"Time of block"`
	TokenID       decimal.Decimal    `bun:"token_id"            comment:"Token ID"`
	Amount        decimal.Decimal    `bun:"amount"              comment:"Transfer amount"`
	Type          types.TransferType `bun:",type:transfer_type" comment:"Transfer type"`
	ContractId    uint64             `bun:"contract_id"         comment:"Contract address id"`
	FromAddressId *uint64            `bun:"from_address_id"     comment:"From address id"`
	ToAddressId   *uint64            `bun:"to_address_id"       comment:"To address id"`
	TxID          uint64             `bun:"tx_id"               comment:"Transaction id"`

	Contract    Contract `bun:"rel:belongs-to,join:contract_id=id"`
	FromAddress *Address `bun:"rel:belongs-to,join:from_address_id=id"`
	ToAddress   *Address `bun:"rel:belongs-to,join:to_address_id=id"`
	Tx          Tx       `bun:"rel:belongs-to,join:tx_id=id"`
}

// TableName -
func (Transfer) TableName() string {
	return "transfer"
}

func (t Transfer) String() string {
	return fmt.Sprintf("%s:%s", t.Contract.String(), t.TokenID.String())
}

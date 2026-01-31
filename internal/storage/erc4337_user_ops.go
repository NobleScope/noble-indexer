package storage

import (
	"context"
	"time"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type ERC4337UserOpsListFilter struct {
	Limit       int
	Offset      int
	Sort        storage.SortOrder
	Height      *uint64
	TxId        *uint64
	BundlerId   *uint64
	PaymasterId *uint64
	Success     *bool
	TimeFrom    time.Time
	TimeTo      time.Time
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IERC4337UserOps interface {
	storage.Table[*ERC4337UserOp]

	Filter(ctx context.Context, filter ERC4337UserOpsListFilter) ([]ERC4337UserOp, error)
}

type ERC4337UserOp struct {
	bun.BaseModel `bun:"erc4337_user_ops" comment:"Table with ERC-4337 user operations."`

	Id                 uint64          `bun:"id,notnull,autoincrement"          comment:"Unique internal identity"`
	Time               time.Time       `bun:"time,notnull"                      comment:"The time of block"`
	Height             pkgTypes.Level  `bun:"height"                            comment:"Block number of the first token occurrence"`
	TxId               uint64          `bun:"tx_id"                             comment:"Transaction identity"`
	Hash               pkgTypes.Hex    `bun:"hash,type:bytea"                   comment:"User operation hash"`
	SenderId           uint64          `bun:"sender_id"                         comment:"Sender address identity"`
	PaymasterId        *uint64         `bun:"paymaster_id"                      comment:"Paymaster identity"`
	BundlerId          uint64          `bun:"bundler_id"                        comment:"Bundler address identity who called handleOps"`
	Nonce              decimal.Decimal `bun:"nonce,type:numeric"                comment:"Nonce"`
	Success            bool            `bun:"success,notnull"                   comment:"Success"`
	ActualGasCost      decimal.Decimal `bun:"actual_gas_cost,type:numeric"      comment:"Actual gas cost"`
	ActualGasUsed      decimal.Decimal `bun:"actual_gas_used,type:numeric"      comment:"Actual gas used"`
	InitCode           pkgTypes.Hex    `bun:"init_code,type:bytea"              comment:"Init code"`
	CallData           pkgTypes.Hex    `bun:"call_data,type:bytea"              comment:"Call data"`
	AccountGasLimits   pkgTypes.Hex    `bun:"account_gas_limits,type:bytea"     comment:"Account gas limits"`
	PreVerificationGas decimal.Decimal `bun:"pre_verification_gas,type:numeric" comment:"Pre verification gas"`
	GasFees            pkgTypes.Hex    `bun:"gas_fees,type:bytea"               comment:"Gas fees"`
	PaymasterAndData   pkgTypes.Hex    `bun:"paymaster_and_data,type:bytea"     comment:"Paymaster and data"`
	Signature          pkgTypes.Hex    `bun:"signature,type:bytea"              comment:"Signature"`

	Tx        Tx       `bun:"rel:belongs-to,join:tx_id=id"`
	Sender    Address  `bun:"rel:belongs-to,join:sender_id=id"`
	Paymaster *Address `bun:"rel:belongs-to,join:paymaster_id=id"`
	Bundler   Address  `bun:"rel:belongs-to,join:bundler_id=id"`
}

// TableName -
func (ERC4337UserOp) TableName() string {
	return "erc4337_user_ops"
}

func (t ERC4337UserOp) String() string {
	return t.Hash.Hex()
}

package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

type ERC4337UserOps struct {
	*postgres.Table[*storage.ERC4337UserOp]
}

// NewERC4337UserOps -
func NewERC4337UserOps(db *database.Bun) *ERC4337UserOps {
	return &ERC4337UserOps{
		Table: postgres.NewTable[*storage.ERC4337UserOp](db),
	}
}

// Filter -
func (u *ERC4337UserOps) Filter(ctx context.Context, filter storage.ERC4337UserOpsListFilter) (userOps []storage.ERC4337UserOp, err error) {
	query := u.DB().NewSelect().Model(&userOps)

	query = erc4337UserOpsListFilter(query, filter)

	outerQuery := u.DB().NewSelect().
		ColumnExpr("user_op.*").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("sender.hash AS sender__hash").
		ColumnExpr("bundler.hash AS bundler__hash").
		ColumnExpr("paymaster.hash AS paymaster__hash").
		TableExpr("(?) AS user_op", query).
		Join("LEFT JOIN tx ON tx.id = user_op.tx_id").
		Join("LEFT JOIN address AS sender ON sender.id = user_op.sender_id").
		Join("LEFT JOIN address AS bundler ON bundler.id = user_op.bundler_id").
		Join("LEFT JOIN address AS paymaster ON paymaster.id = user_op.paymaster_id")

	outerQuery = sortTimeIDScope(outerQuery, filter.Sort)

	err = outerQuery.Scan(ctx, &userOps)
	return
}

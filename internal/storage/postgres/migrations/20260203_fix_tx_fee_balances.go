package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upFixTxFeeBalances, downFixTxFeeBalances)
}

func upFixTxFeeBalances(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `
		WITH diffs AS (
			SELECT from_address_id AS id, SUM(fee - gas_used * effective_gas_price) AS diff
			FROM public."tx"
			WHERE cumulative_gas_used != gas_used
			GROUP BY from_address_id
		)
		UPDATE public."balance"
		SET "value" = "value" + diffs.diff
		FROM diffs
		WHERE diffs.id = balance.id
	`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE public."tx"
		SET "fee" = gas_used * effective_gas_price
		WHERE cumulative_gas_used != gas_used
	`); err != nil {
		return err
	}

	return nil
}

func downFixTxFeeBalances(ctx context.Context, db *bun.DB) error {
	return nil
}

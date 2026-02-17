package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upFixContractCreateBalances, downFixContractCreateBalances)
}

func upFixContractCreateBalances(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		WITH create_values AS (
			SELECT t.contract_id AS id, SUM(t.amount) AS total_value
			FROM public."trace" t
			WHERE t.type IN ('create', 'create2')
			  AND t.amount > 0
			  AND t.contract_id IS NOT NULL
			GROUP BY t.contract_id
		)
		UPDATE public."balance"
		SET "value" = "value" + create_values.total_value
		FROM create_values
		WHERE create_values.id = balance.id
	`)
	return err
}

func downFixContractCreateBalances(ctx context.Context, db *bun.DB) error {
	return nil
}

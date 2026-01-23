package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upAddTotalContractsColumns, downAddTotalContractsColumns)
}

func upAddTotalContractsColumns(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE public."state" ADD COLUMN IF NOT EXISTS "total_contracts" int8 NOT NULL DEFAULT 0`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `COMMENT ON COLUMN public."state"."total_contracts" IS 'Contracts count'`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `ALTER TABLE public."state" ADD COLUMN IF NOT EXISTS "total_tokens" int8 NOT NULL DEFAULT 0`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `COMMENT ON COLUMN public."state"."total_tokens" IS 'Tokens count'`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE public."state"
		SET total_contracts = (SELECT COUNT(*) FROM public."contract")
	`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE public."state"
		SET total_tokens = (SELECT COUNT(*) FROM public."token")
	`); err != nil {
		return err
	}

	return nil
}

func downAddTotalContractsColumns(ctx context.Context, db *bun.DB) error {
	return nil
}

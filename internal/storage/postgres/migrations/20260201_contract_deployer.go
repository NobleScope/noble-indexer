package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upAddContractDeployer, downAddContractDeployer)
}

func upAddContractDeployer(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE public."contract" ADD COLUMN IF NOT EXISTS "deployer_id" bigint`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `COMMENT ON COLUMN public."contract"."deployer_id" IS 'Deployer address ID'`); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE public."contract" AS c
		SET "deployer_id" = t."from_address_id"
		FROM public."tx" AS t
		WHERE c."tx_id" = t."id"
		  AND c."deployer_id" IS NULL
	`); err != nil {
		return err
	}

	return nil
}

func downAddContractDeployer(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE public."contract" DROP COLUMN IF EXISTS "deployer_id"`); err != nil {
		return err
	}
	return nil
}

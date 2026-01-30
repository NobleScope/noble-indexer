package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upAddTokenLogo, downAddTokenLogo)
}

func upAddTokenLogo(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE public."token" ADD COLUMN IF NOT EXISTS "logo" text`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `COMMENT ON COLUMN public."token"."logo" IS 'Logo URL'`); err != nil {
		return err
	}

	return nil
}

func downAddTokenLogo(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE public."token" DROP COLUMN IF EXISTS "logo"`); err != nil {
		return err
	}
	return nil
}

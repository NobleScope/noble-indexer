package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upFixAddressIsContract, downFixAddressIsContract)
}

func upFixAddressIsContract(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		UPDATE public."address" AS a
		SET "is_contract" = true
		FROM public."contract" AS c
		WHERE a."id" = c."id"
		  AND a."is_contract" = false
	`)
	return err
}

func downFixAddressIsContract(ctx context.Context, db *bun.DB) error {
	return nil
}

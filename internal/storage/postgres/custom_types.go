package postgres

import (
	"context"
	"database/sql"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/dipdup-net/go-lib/database"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

const (
	createTypeQuery = `DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = ?) THEN
			CREATE TYPE ? AS ENUM (?);
		END IF;
	END$$;`
)

func createTypes(ctx context.Context, conn *database.Bun) error {
	log.Info().Msg("creating custom types...")
	return conn.DB().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			createTypeQuery,
			"tx_type",
			bun.Safe("tx_type"),
			bun.In(types.TxTypeValues()),
		); err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			createTypeQuery,
			"tx_status",
			bun.Safe("tx_status"),
			bun.In(types.TxStatusValues()),
		); err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			createTypeQuery,
			"transfer_type",
			bun.Safe("transfer_type"),
			bun.In(types.TransferTypeValues()),
		); err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			createTypeQuery,
			"token_type",
			bun.Safe("token_type"),
			bun.In(types.TokenTypeValues()),
		); err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			createTypeQuery,
			"metadata_status",
			bun.Safe("metadata_status"),
			bun.In(types.MetadataStatusValues()),
		); err != nil {
			return err
		}

		return nil
	})
}

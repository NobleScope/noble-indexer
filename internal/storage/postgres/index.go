package postgres

import (
	"context"
	"database/sql"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

func createIndices(ctx context.Context, conn *database.Bun) error {
	log.Info().Msg("creating indexes...")
	return conn.DB().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		// Block
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Block)(nil)).
			Index("block_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Block)(nil)).
			Index("block_hash_idx").
			Column("hash").
			Using("HASH").
			Exec(ctx); err != nil {
			return err
		}

		// BlockStats
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.BlockStats)(nil)).
			Index("block_stats_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}

		// Tx
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Tx)(nil)).
			Index("tx_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Tx)(nil)).
			Index("tx_hash_idx").
			Column("hash").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Tx)(nil)).
			Index("tx_status_idx").
			Column("status").
			Exec(ctx); err != nil {
			return err
		}

		// Log
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Log)(nil)).
			Index("log_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Log)(nil)).
			Index("log_tx_id_idx").
			Column("tx_id").
			Exec(ctx); err != nil {
			return err
		}

		// Contract
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Contract)(nil)).
			Index("contract_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Contract)(nil)).
			Index("contract_metadata_link_idx").
			Column("metadata_link").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Contract)(nil)).
			Index("contract_tx_id_idx").
			Column("tx_id").
			Exec(ctx); err != nil {
			return err
		}

		// Trace
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Trace)(nil)).
			Index("trace_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Trace)(nil)).
			Index("trace_tx_id_idx").
			Column("tx_id").
			Exec(ctx); err != nil {
			return err
		}

		// Address
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Address)(nil)).
			Index("address_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Address)(nil)).
			Index("address_address_idx").
			Column("address").
			Exec(ctx); err != nil {
			return err
		}

		// Transfer
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Transfer)(nil)).
			Index("transfer_height_idx").
			Column("height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Transfer)(nil)).
			Index("transfer_tx_id_idx").
			Column("tx_id").
			Exec(ctx); err != nil {
			return err
		}

		// Token
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Token)(nil)).
			Index("token_symbol_idx").
			ColumnExpr("symbol gin_trgm_ops").
			Using("GIN").
			Exec(ctx); err != nil {
			return err
		}

		// Source
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Source)(nil)).
			Index("source_contract_id_idx").
			Column("contract_id").
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

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
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Tx)(nil)).
			Index("tx_type_idx").
			Column("type").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Tx)(nil)).
			Index("tx_from_address_id_idx").
			Column("from_address_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Tx)(nil)).
			Index("tx_to_address_id_idx").
			Column("to_address_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Tx)(nil)).
			Index("tx_height_time_id_idx").
			Column("height", "time", "id").
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
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Log)(nil)).
			Index("log_address_id_idx").
			Column("address_id").
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
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Trace)(nil)).
			Index("trace_from_address_id_idx").
			Column("from_address_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Trace)(nil)).
			Index("trace_to_address_id_idx").
			Column("to_address_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Trace)(nil)).
			Index("trace_contract_id_idx").
			Column("contract_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Trace)(nil)).
			Index("trace_type_idx").
			Column("type").
			Exec(ctx); err != nil {
			return err
		}

		// Address
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Address)(nil)).
			Index("address_first_height_idx").
			Column("first_height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Address)(nil)).
			Index("address_hash_idx").
			Column("hash").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Address)(nil)).
			Index("address_is_contract_idx").
			Column("is_contract").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Address)(nil)).
			Index("address_last_height_idx").
			Column("last_height").
			Using("BRIN").
			Exec(ctx); err != nil {
			return err
		}

		// Balance
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Balance)(nil)).
			Index("balance_value_idx").
			Column("value").
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
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Transfer)(nil)).
			Index("transfer_token_id_idx").
			Column("token_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Transfer)(nil)).
			Index("transfer_type_idx").
			Column("type").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Transfer)(nil)).
			Index("transfer_contract_id_idx").
			Column("contract_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Transfer)(nil)).
			Index("transfer_from_address_id_idx").
			Column("from_address_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Transfer)(nil)).
			Index("transfer_to_address_id_idx").
			Column("to_address_id").
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
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Token)(nil)).
			Index("token_name_idx").
			ColumnExpr("name gin_trgm_ops").
			Using("GIN").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.Token)(nil)).
			Index("token_type_idx").
			Column("type").
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

		// Index for proxy contracts by status and height
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.ProxyContract)(nil)).
			Index("proxy_contract_status_height_idx").
			Column("status", "height").
			Exec(ctx); err != nil {
			return err
		}
		// Index for implementation lookup
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.ProxyContract)(nil)).
			Index("proxy_contract_implementation_idx").
			Column("implementation_id").
			Exec(ctx); err != nil {
			return err
		}

		// Verification files
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.VerificationFile)(nil)).
			Index("verification_file_verification_task_id_idx").
			Column("verification_task_id").
			Exec(ctx); err != nil {
			return err
		}

		// Verification task
		if _, err := tx.NewCreateIndex().
			IfNotExists().
			Model((*storage.VerificationTask)(nil)).
			Index("verification_task_contract_id_idx").
			Column("contract_id").
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

package postgres

import (
	"context"

	models "github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// Storage -
type Storage struct {
	*postgres.Storage

	cfg        config.Database
	scriptsDir string

	Blocks     models.IBlock
	BlockStats models.IBlockStats
	Tx         models.ITx
	Logs       models.ILog
	Addresses  models.IAddress
	Contracts  models.IContract
	State      models.IState
}

// Create -
func Create(ctx context.Context, cfg config.Database, scriptsDir string) (Storage, error) {
	init := initDatabase

	strg, err := postgres.Create(ctx, cfg, init)
	if err != nil {
		return Storage{}, err
	}

	s := Storage{
		cfg:        cfg,
		scriptsDir: scriptsDir,
		Storage:    strg,
		Blocks:     NewBlock(strg.Connection()),
		BlockStats: NewBlockStats(strg.Connection()),
		Logs:       NewLog(strg.Connection()),
		Tx:         NewTx(strg.Connection()),
		Addresses:  NewAddress(strg.Connection()),
		Contracts:  NewContract(strg.Connection()),
		State:      NewState(strg.Connection()),
	}

	if err := s.createScripts(ctx, "functions", false); err != nil {
		return s, errors.Wrap(err, "creating views")
	}
	if err := s.createScripts(ctx, "views", true); err != nil {
		return s, errors.Wrap(err, "creating views")
	}

	return s, nil
}

func initDatabase(ctx context.Context, conn *database.Bun) error {
	if err := createExtensions(ctx, conn); err != nil {
		return errors.Wrap(err, "create extensions")
	}
	if err := createTypes(ctx, conn); err != nil {
		return errors.Wrap(err, "creating custom types")
	}

	if err := database.CreateTables(ctx, conn, models.Models...); err != nil {
		if err := conn.Close(); err != nil {
			return err
		}
		return err
	}

	if err := database.MakeComments(ctx, conn, models.Models...); err != nil {
		if err := conn.Close(); err != nil {
			return err
		}
		return errors.Wrap(err, "make comments")
	}

	if err := createHypertables(ctx, conn); err != nil {
		if err := conn.Close(); err != nil {
			return err
		}
		return errors.Wrap(err, "create hypertables")
	}

	return createIndices(ctx, conn)
}

func createHypertables(ctx context.Context, conn *database.Bun) error {
	return conn.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, model := range []storage.Model{
			&models.Block{},
			&models.BlockStats{},
			&models.Tx{},
			&models.Log{},
		} {
			if _, err := tx.ExecContext(ctx,
				`SELECT create_hypertable(?, 'time', chunk_time_interval => INTERVAL '1 month', if_not_exists => TRUE);`,
				model.TableName(),
			); err != nil {
				return err
			}
		}

		return nil
	})
}

func createExtensions(ctx context.Context, conn *database.Bun) error {
	return conn.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pg_trgm;")
		return err
	})
}

func (s Storage) Close() error {
	if err := s.Storage.Close(); err != nil {
		return err
	}
	return nil
}

package postgres

import (
	"context"

	models "github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres/migrations"
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// Storage -
type Storage struct {
	*postgres.Storage

	cfg        config.Database
	scriptsDir string

	Blocks         models.IBlock
	BlockStats     models.IBlockStats
	Tx             models.ITx
	Transfer       models.ITransfer
	Token          models.IToken
	TokenBalance   models.ITokenBalance
	Trace          models.ITrace
	Logs           models.ILog
	Addresses      models.IAddress
	Contracts      models.IContract
	ProxyContracts models.IProxyContract
	Sources        models.ISource
	State          models.IState
	Search         models.ISearch
	ERC4337UserOps models.IERC4337UserOps
	Notificator    *Notificator
}

// Create -
func Create(ctx context.Context, cfg config.Database, scriptsDir string, withMigrations bool) (Storage, error) {
	init := initDatabase
	if withMigrations {
		init = initDatabaseWithMigrations
	}
	strg, err := postgres.Create(ctx, cfg, init)
	if err != nil {
		return Storage{}, err
	}

	s := Storage{
		cfg:            cfg,
		scriptsDir:     scriptsDir,
		Storage:        strg,
		Blocks:         NewBlock(strg.Connection()),
		BlockStats:     NewBlockStats(strg.Connection()),
		Logs:           NewLog(strg.Connection()),
		Tx:             NewTx(strg.Connection()),
		Transfer:       NewTransfer(strg.Connection()),
		Token:          NewToken(strg.Connection()),
		TokenBalance:   NewTokenBalance(strg.Connection()),
		Trace:          NewTrace(strg.Connection()),
		Addresses:      NewAddress(strg.Connection()),
		Contracts:      NewContract(strg.Connection()),
		ProxyContracts: NewProxyContract(strg.Connection()),
		Sources:        NewSource(strg.Connection()),
		State:          NewState(strg.Connection()),
		Search:         NewSearch(strg.Connection()),
		ERC4337UserOps: NewERC4337UserOps(strg.Connection()),
		Notificator:    NewNotificator(cfg, strg.Connection().DB()),
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
		if errClose := conn.Close(); errClose != nil {
			return errors.Wrap(errClose, err.Error())
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

func initDatabaseWithMigrations(ctx context.Context, conn *database.Bun) error {
	exists, err := checkTablesExists(ctx, conn)
	if err != nil {
		return errors.Wrap(err, "check table exists")
	}

	if exists {
		if err := migrateDatabase(ctx, conn); err != nil {
			return errors.Wrap(err, "migrate database")
		}
	}

	return initDatabase(ctx, conn)
}

func createHypertables(ctx context.Context, conn *database.Bun) error {
	return conn.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, model := range []storage.Model{
			&models.Block{},
			&models.BlockStats{},
			&models.Tx{},
			&models.Transfer{},
			&models.Log{},
			&models.ERC4337UserOp{},
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

func migrateDatabase(ctx context.Context, db *database.Bun) error {
	migrator := migrate.NewMigrator(db.DB(), migrations.Migrations)
	if err := migrator.Init(ctx); err != nil {
		return err
	}
	if err := migrator.Lock(ctx); err != nil {
		return err
	}
	defer migrator.Unlock(ctx) //nolint:errcheck

	_, err := migrator.Migrate(ctx)
	return err
}

func (s Storage) CreateListener() models.Listener {
	return NewNotificator(s.cfg, s.Notificator.db)
}

func (s Storage) Close() error {
	if err := s.Storage.Close(); err != nil {
		return err
	}
	return nil
}

func checkTablesExists(ctx context.Context, db *database.Bun) (bool, error) {
	var exists bool
	err := db.DB().NewRaw(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE  table_schema = 'public'
		AND    table_name   = 'state'
	)`).Scan(ctx, &exists)
	return exists, err
}

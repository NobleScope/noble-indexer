package contract_verifier

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/pkg/errors"
)

type Module struct {
	modules.BaseModule

	pg         postgres.Storage
	storage    sdk.Transactable
	syncPeriod time.Duration
	retryDelay time.Duration
	cfg        config.Config
}

func NewModule(pg postgres.Storage, cfg config.Config) *Module {
	module := &Module{
		BaseModule: modules.New("contract_verifier"),
		pg:         pg,
		storage:    pg.Transactable,
		cfg:        cfg,
		syncPeriod: time.Second * time.Duration(cfg.ContractVerifier.SyncPeriod),
	}

	return module
}

func (m *Module) Close() error {
	m.Log.Info().Msg("closing module...")
	m.G.Wait()

	return nil
}

func (m *Module) Start(ctx context.Context) {
	m.Log.Info().Msg("starting module...")
	m.G.GoCtx(ctx, m.receive)
}

func (m *Module) receive(ctx context.Context) {
	if err := m.sync(ctx); err != nil {
		m.Log.Err(err).Msg("sync")
	}

	ticker := time.NewTicker(m.syncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.sync(ctx); err != nil {
				m.Log.Err(err).Msg("sync")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Module) sync(ctx context.Context) error {
	task, err := m.pg.VerificationTasks.Latest(ctx)
	if err != nil {
		return errors.Wrap(err, "get verification task")
	}
	if task.Id == 0 {
		m.Log.Info().Msg("no tasks for contract verification")
		return nil
	}
	m.Log.Info().
		Uint64("contract id", task.ContractId).
		Time("creation task time", task.CreationTime).
		Msg("get verification task")

	m.verify()
	//if err := m.save(ctx, contracts, sources); err != nil {
	//	m.Log.Err(err).Msg("save")
	//}

	return err
}

func (m *Module) save(ctx context.Context, contracts []*storage.Contract, sources []*storage.Source) error {
	tx, err := postgres.BeginTransaction(ctx, m.storage)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	err = tx.SaveContracts(ctx, contracts...)
	if err != nil {
		return errors.Wrap(err, "save contracts")
	}
	m.Log.Info().Int("count", len(contracts)).Msg("contract metadata updated")

	err = tx.SaveSources(ctx, sources...)
	if err != nil {
		return errors.Wrap(err, "save contract sources")
	}
	m.Log.Info().Int("count", len(sources)).Msg("contract sources saved")

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	return nil
}

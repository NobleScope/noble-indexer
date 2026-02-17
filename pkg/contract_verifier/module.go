package contract_verifier

import (
	"context"
	"database/sql"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
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
		if errors.Is(err, sql.ErrNoRows) {
			m.Log.Info().Msg("no tasks for contract verification")
			return nil
		}
		return errors.Wrap(err, "get verification task")
	}

	m.Log.Info().
		Uint64("contract_id", task.ContractId).
		Time("creation_time", task.CreationTime).
		Msg("processing verification task")

	files, err := m.pg.VerificationFiles.ByTaskId(ctx, task.Id)
	if err != nil {
		return errors.Wrap(err, "get verification files")
	}

	if len(files) == 0 {
		m.Log.Warn().Uint64("task_id", task.Id).Msg("no files found for verification task")
		if err := m.handleVerificationFailure(ctx, task, "no files found for verification task"); err != nil {
			return err
		}
		return nil
	}

	result, verifyErr := m.verify(ctx, task, files)
	if verifyErr != nil {
		m.Log.Err(verifyErr).Msg("verification failed")
		if err := m.handleVerificationFailure(ctx, task, verifyErr.Error()); err != nil {
			return err
		}
		return nil
	}

	m.Log.Info().
		Uint64("contract_id", task.ContractId).
		Msg("contract verified successfully")

	if err := m.handleVerificationSuccess(ctx, task, files, result); err != nil {
		return err
	}

	return nil
}

func (m *Module) handleVerificationSuccess(ctx context.Context, task storage.VerificationTask, files []storage.VerificationFile, result *VerificationResult) error {
	tx, err := postgres.BeginTransaction(ctx, m.storage)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	contract, err := m.pg.Contracts.GetByID(ctx, task.ContractId)
	if err != nil {
		return errors.Wrap(err, "get contract")
	}

	contract.Verified = true
	contract.ABI = result.ABI
	contract.CompilerVersion = result.CompilerVersion
	contract.Language = result.Language
	if task.OptimizationEnabled != nil {
		contract.OptimizerEnabled = *task.OptimizationEnabled
	}

	sources := make([]*storage.Source, 0, len(files))
	for _, file := range files {
		sources = append(sources, &storage.Source{
			Name:       file.Name,
			License:    task.LicenseType.String(),
			Content:    string(file.File),
			ContractId: task.ContractId,
		})
	}

	err = tx.SaveContracts(ctx, contract)
	if err != nil {
		return errors.Wrap(err, "save contract")
	}

	err = tx.SaveSources(ctx, sources...)
	if err != nil {
		return errors.Wrap(err, "save contract sources")
	}

	task.Status = types.VerificationStatusSuccess
	if err := tx.UpdateVerificationTask(ctx, &task); err != nil {
		return errors.Wrap(err, "update task status")
	}

	if err := tx.DeleteVerificationFiles(ctx, task.Id); err != nil {
		return errors.Wrap(err, "delete verification files")
	}

	if err := m.updateState(ctx); err != nil {
		m.Log.Err(err).Msg("update state")
	}

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	return nil
}

func (m *Module) handleVerificationFailure(ctx context.Context, task storage.VerificationTask, errorMsg string) error {
	tx, err := postgres.BeginTransaction(ctx, m.storage)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	task.Status = types.VerificationStatusFailed
	task.Error = errorMsg
	if err := tx.UpdateVerificationTask(ctx, &task); err != nil {
		return errors.Wrap(err, "update task status")
	}

	if err := tx.DeleteVerificationFiles(ctx, task.Id); err != nil {
		return errors.Wrap(err, "delete verification files")
	}

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	return nil
}

func (m *Module) updateState(ctx context.Context) error {
	state, err := m.pg.State.ByName(ctx, m.cfg.Indexer.Name)
	if err != nil {
		return errors.Wrap(err, "get state")
	}

	state.TotalVerifiedContracts++

	if err := m.pg.State.Update(ctx, &state); err != nil {
		return errors.Wrap(err, "update state")
	}

	return nil
}

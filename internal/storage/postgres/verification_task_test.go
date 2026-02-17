package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage/types"
)

func (s *StorageTestSuite) TestVerificationTaskLatest() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	task, err := s.storage.VerificationTasks.Latest(ctx)
	s.Require().NoError(err)

	s.Require().EqualValues(1, task.Id)
	s.Require().EqualValues(types.VerificationStatusNew, task.Status)
	s.Require().EqualValues(3, task.ContractId)
	s.Require().EqualValues("Token", task.ContractName)
	s.Require().EqualValues("v0.8.20+commit.a1b79de6", task.CompilerVersion)
	s.Require().EqualValues(types.Mit, task.LicenseType)
	s.Require().NotNil(task.OptimizationEnabled)
	s.Require().True(*task.OptimizationEnabled)
	s.Require().NotNil(task.OptimizationRuns)
	s.Require().EqualValues(200, *task.OptimizationRuns)
	s.Require().False(task.ViaIR)
	s.Require().Empty(task.Error)
}

func (s *StorageTestSuite) TestVerificationTaskLatestSkipsNonNew() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	task, err := s.storage.VerificationTasks.Latest(ctx)
	s.Require().NoError(err)

	s.Require().EqualValues(types.VerificationStatusNew, task.Status)
	s.Require().NotEqualValues(types.VerificationStatusSuccess, task.Status)
	s.Require().NotEqualValues(types.VerificationStatusFailed, task.Status)
	s.Require().NotEqualValues(types.VerificationStatusPending, task.Status)
}

func (s *StorageTestSuite) TestVerificationTaskByContractId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tasks, err := s.storage.VerificationTasks.ByContractId(ctx, 3)
	s.Require().NoError(err)
	s.Require().Len(tasks, 1)

	task := tasks[0]
	s.Require().EqualValues(1, task.Id)
	s.Require().EqualValues(types.VerificationStatusNew, task.Status)
	s.Require().EqualValues(3, task.ContractId)
	s.Require().EqualValues("Token", task.ContractName)
	s.Require().EqualValues("v0.8.20+commit.a1b79de6", task.CompilerVersion)
	s.Require().EqualValues(types.Mit, task.LicenseType)
}

func (s *StorageTestSuite) TestVerificationTaskByContractIdSuccessStatus() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tasks, err := s.storage.VerificationTasks.ByContractId(ctx, 4)
	s.Require().NoError(err)
	s.Require().Len(tasks, 1)

	task := tasks[0]
	s.Require().EqualValues(2, task.Id)
	s.Require().EqualValues(types.VerificationStatusSuccess, task.Status)
	s.Require().EqualValues(4, task.ContractId)
	s.Require().EqualValues("Proxy", task.ContractName)
	s.Require().EqualValues("v0.8.19+commit.7dd6d404", task.CompilerVersion)
	s.Require().EqualValues(types.Apache20, task.LicenseType)
	s.Require().False(task.CompletionTime.IsZero())
}

func (s *StorageTestSuite) TestVerificationTaskByContractIdFailed() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tasks, err := s.storage.VerificationTasks.ByContractId(ctx, 5)
	s.Require().NoError(err)
	s.Require().Len(tasks, 1)

	task := tasks[0]
	s.Require().EqualValues(3, task.Id)
	s.Require().EqualValues(types.VerificationStatusFailed, task.Status)
	s.Require().EqualValues("Implementation", task.ContractName)
	s.Require().NotNil(task.OptimizationEnabled)
	s.Require().True(*task.OptimizationEnabled)
	s.Require().NotNil(task.OptimizationRuns)
	s.Require().EqualValues(500, *task.OptimizationRuns)
	s.Require().NotNil(task.EVMVersion)
	s.Require().EqualValues(types.Shanghai, *task.EVMVersion)
	s.Require().True(task.ViaIR)
	s.Require().EqualValues("bytecode verification failed: main parts do not match", task.Error)
}

func (s *StorageTestSuite) TestVerificationTaskByContractIdPending() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tasks, err := s.storage.VerificationTasks.ByContractId(ctx, 6)
	s.Require().NoError(err)
	s.Require().Len(tasks, 1)

	task := tasks[0]
	s.Require().EqualValues(4, task.Id)
	s.Require().EqualValues(types.VerificationStatusPending, task.Status)
	s.Require().EqualValues("Vault", task.ContractName)
	s.Require().EqualValues("v0.8.24+commit.e11b9ed9", task.CompilerVersion)
	s.Require().EqualValues(types.Unlicense, task.LicenseType)
	s.Require().NotNil(task.OptimizationEnabled)
	s.Require().True(*task.OptimizationEnabled)
	s.Require().NotNil(task.OptimizationRuns)
	s.Require().EqualValues(1000, *task.OptimizationRuns)
	s.Require().NotNil(task.EVMVersion)
	s.Require().EqualValues(types.Cancun, *task.EVMVersion)
	s.Require().False(task.ViaIR)
}

func (s *StorageTestSuite) TestVerificationTaskByContractIdNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tasks, err := s.storage.VerificationTasks.ByContractId(ctx, 999)
	s.Require().NoError(err)
	s.Require().Len(tasks, 0)
}

func (s *StorageTestSuite) TestVerificationTaskGetByID() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	task, err := s.storage.VerificationTasks.GetByID(ctx, 1)
	s.Require().NoError(err)
	s.Require().EqualValues(1, task.Id)
	s.Require().EqualValues(types.VerificationStatusNew, task.Status)
	s.Require().EqualValues(3, task.ContractId)
	s.Require().EqualValues("Token", task.ContractName)
}

func (s *StorageTestSuite) TestVerificationTaskGetByIDNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.storage.VerificationTasks.GetByID(ctx, 999)
	s.Require().Error(err)
	s.Require().ErrorIs(err, sql.ErrNoRows)
}

func (s *StorageTestSuite) TestVerificationTaskNilOptimization() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	task, err := s.storage.VerificationTasks.GetByID(ctx, 2)
	s.Require().NoError(err)
	s.Require().NotNil(task.OptimizationEnabled)
	s.Require().False(*task.OptimizationEnabled)
	s.Require().Nil(task.OptimizationRuns)
	s.Require().Nil(task.EVMVersion)
}

func (s *StorageTestSuite) TestVerificationTaskWithEVMVersionAndViaIR() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	task, err := s.storage.VerificationTasks.GetByID(ctx, 3)
	s.Require().NoError(err)
	s.Require().NotNil(task.EVMVersion)
	s.Require().EqualValues(types.Shanghai, *task.EVMVersion)
	s.Require().True(task.ViaIR)
}

func (s *StorageTestSuite) TestVerificationTaskAllStatuses() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	expectedStatuses := map[uint64]types.VerificationTaskStatus{
		1: types.VerificationStatusNew,
		2: types.VerificationStatusSuccess,
		3: types.VerificationStatusFailed,
		4: types.VerificationStatusPending,
	}

	for id, expectedStatus := range expectedStatuses {
		task, err := s.storage.VerificationTasks.GetByID(ctx, id)
		s.Require().NoError(err, "task_id: %d", id)
		s.Require().EqualValues(expectedStatus, task.Status, "task_id: %d", id)
	}
}

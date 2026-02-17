package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
)

func (s *StorageTestSuite) TestVerificationFileByTaskIdMultipleFiles() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	files, err := s.storage.VerificationFiles.ByTaskId(ctx, 1)
	s.Require().NoError(err)
	s.Require().Len(files, 2)

	s.Require().EqualValues(1, files[0].Id)
	s.Require().EqualValues("Token.sol", files[0].Name)
	s.Require().EqualValues(uint64(1), files[0].VerificationTaskId)
	s.Require().NotEmpty(files[0].File)

	s.Require().EqualValues(2, files[1].Id)
	s.Require().EqualValues("IERC20.sol", files[1].Name)
	s.Require().EqualValues(uint64(1), files[1].VerificationTaskId)
}

func (s *StorageTestSuite) TestVerificationFileByTaskIdSingleFile() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	files, err := s.storage.VerificationFiles.ByTaskId(ctx, 2)
	s.Require().NoError(err)
	s.Require().Len(files, 1)

	s.Require().EqualValues(3, files[0].Id)
	s.Require().EqualValues("Proxy.sol", files[0].Name)
	s.Require().EqualValues(uint64(2), files[0].VerificationTaskId)
}

func (s *StorageTestSuite) TestVerificationFileByTaskIdFailedTask() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	files, err := s.storage.VerificationFiles.ByTaskId(ctx, 3)
	s.Require().NoError(err)
	s.Require().Len(files, 1)

	s.Require().EqualValues(4, files[0].Id)
	s.Require().EqualValues("Implementation.sol", files[0].Name)
}

func (s *StorageTestSuite) TestVerificationFileByTaskIdTwoFiles() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	files, err := s.storage.VerificationFiles.ByTaskId(ctx, 4)
	s.Require().NoError(err)
	s.Require().Len(files, 2)

	s.Require().EqualValues("Vault.sol", files[0].Name)
	s.Require().EqualValues("VaultStorage.sol", files[1].Name)

	for _, f := range files {
		s.Require().EqualValues(uint64(4), f.VerificationTaskId)
	}
}

func (s *StorageTestSuite) TestVerificationFileByTaskIdNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	files, err := s.storage.VerificationFiles.ByTaskId(ctx, 999)
	s.Require().NoError(err)
	s.Require().Len(files, 0)
}

func (s *StorageTestSuite) TestVerificationFileGetByID() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	file, err := s.storage.VerificationFiles.GetByID(ctx, 1)
	s.Require().NoError(err)
	s.Require().EqualValues(1, file.Id)
	s.Require().EqualValues("Token.sol", file.Name)
	s.Require().EqualValues(uint64(1), file.VerificationTaskId)
	s.Require().NotEmpty(file.File)
}

func (s *StorageTestSuite) TestVerificationFileGetByIDNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.storage.VerificationFiles.GetByID(ctx, 999)
	s.Require().Error(err)
	s.Require().ErrorIs(err, sql.ErrNoRows)
}

func (s *StorageTestSuite) TestVerificationFileBulkSaveEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	err := s.storage.VerificationFiles.BulkSave(ctx)
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestVerificationFileContentNotEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	files, err := s.storage.VerificationFiles.ByTaskId(ctx, 1)
	s.Require().NoError(err)
	s.Require().Len(files, 2)

	for _, f := range files {
		s.Require().NotEmpty(f.File)
		s.Require().Contains(string(f.File), "pragma solidity")
	}
}

func (s *StorageTestSuite) TestVerificationFileCountsByTask() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	testCases := []struct {
		taskId        uint64
		expectedCount int
		expectedNames []string
	}{
		{1, 2, []string{"Token.sol", "IERC20.sol"}},
		{2, 1, []string{"Proxy.sol"}},
		{3, 1, []string{"Implementation.sol"}},
		{4, 2, []string{"Vault.sol", "VaultStorage.sol"}},
	}

	for _, tc := range testCases {
		files, err := s.storage.VerificationFiles.ByTaskId(ctx, tc.taskId)
		s.Require().NoError(err, "task_id: %d", tc.taskId)
		s.Require().Len(files, tc.expectedCount, "task_id: %d", tc.taskId)

		for i, f := range files {
			s.Require().EqualValues(tc.expectedNames[i], f.Name, "task_id: %d, index: %d", tc.taskId, i)
		}
	}
}

func (s *StorageTestSuite) TestVerificationFileZBulkSaveAndRead() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	newFiles := []*storage.VerificationFile{
		{
			Name:               "NewContract.sol",
			File:               []byte("pragma solidity ^0.8.20;\ncontract NewContract {}"),
			VerificationTaskId: 1,
		},
		{
			Name:               "Helper.sol",
			File:               []byte("pragma solidity ^0.8.20;\nlibrary Helper {}"),
			VerificationTaskId: 1,
		},
	}

	err := s.storage.VerificationFiles.BulkSave(ctx, newFiles...)
	s.Require().NoError(err)

	s.Require().NotZero(newFiles[0].Id)
	s.Require().NotZero(newFiles[1].Id)

	files, err := s.storage.VerificationFiles.ByTaskId(ctx, 1)
	s.Require().NoError(err)
	s.Require().Len(files, 4)
}

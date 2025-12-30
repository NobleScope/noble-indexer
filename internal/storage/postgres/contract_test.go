package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestContractByHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contract, err := s.storage.Contracts.ByHash(ctx, pkgTypes.MustDecodeHex("0x30f055506ba543ea0942dc8ca03f596ab75bc879"))
	s.Require().NoError(err)
	s.Require().EqualValues(3, contract.Id)
	s.Require().EqualValues(100, contract.Height)
	s.Require().True(contract.Verified)
	s.Require().NotNil(contract.TxId)
	s.Require().EqualValues(1, *contract.TxId)
	s.Require().EqualValues("v0.8.20+commit.a1b79de6", contract.CompilerVersion)
	s.Require().EqualValues("Solidity", contract.Language)
	s.Require().True(contract.OptimizerEnabled)
	s.Require().Len(contract.Tags, 2)
	s.Require().Contains(contract.Tags, "ERC20")
	s.Require().Contains(contract.Tags, "Pausable")
	s.Require().EqualValues(types.Success, contract.Status)
	s.Require().EqualValues(0, contract.RetryCount)

	// Check Address relation
	s.Require().EqualValues(3, contract.Address.Id)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", contract.Address.Hash.Hex())

	// Check Tx relation
	s.Require().NotNil(contract.Tx)
	s.Require().EqualValues("0x90f5df4e03620cc55d3ea295bf8826f84465065340cb6d0d095166dd2465f283", contract.Tx.Hash.Hex())

	// No proxy implementation for this contract
	s.Require().Nil(contract.Implementation)
}

func (s *StorageTestSuite) TestContractByHashWithProxy() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contract, err := s.storage.Contracts.ByHash(ctx, pkgTypes.MustDecodeHex("0x60f055506ba543ea0942dc8ca03f596ab75bc882"))
	s.Require().NoError(err)
	s.Require().EqualValues(6, contract.Id)
	s.Require().EqualValues(400, contract.Height)
	s.Require().True(contract.Verified)

	// Check proxy implementation
	s.Require().NotNil(contract.Implementation)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", contract.Implementation.Hex())
}

func (s *StorageTestSuite) TestContractByHashNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.storage.Contracts.ByHash(ctx, pkgTypes.MustDecodeHex("0x0000000000000000000000000000000000000001"))
	s.Require().Error(err)
	s.Require().ErrorIs(err, sql.ErrNoRows)
}

func (s *StorageTestSuite) TestContractListWithTx() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.ListWithTx(ctx, storage.ContractListFilter{
		Limit:      10,
		Offset:     0,
		Sort:       sdk.SortOrderAsc,
		SortField:  "id",
		IsVerified: false,
		TxId:       nil,
	})
	s.Require().NoError(err)
	s.Require().Len(contracts, 5)

	// Check first contract
	s.Require().EqualValues(3, contracts[0].Id)
	s.Require().EqualValues(100, contracts[0].Height)
	s.Require().True(contracts[0].Verified)
	s.Require().NotNil(contracts[0].Tx)
	s.Require().EqualValues("0x90f5df4e03620cc55d3ea295bf8826f84465065340cb6d0d095166dd2465f283", contracts[0].Tx.Hash.Hex())

	// Check last contract
	s.Require().EqualValues(7, contracts[4].Id)
	s.Require().EqualValues(500, contracts[4].Height)
	s.Require().False(contracts[4].Verified)
}

func (s *StorageTestSuite) TestContractListWithTxVerifiedOnly() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.ListWithTx(ctx, storage.ContractListFilter{
		Limit:      10,
		Offset:     0,
		Sort:       sdk.SortOrderAsc,
		SortField:  "id",
		IsVerified: true,
		TxId:       nil,
	})
	s.Require().NoError(err)
	s.Require().Len(contracts, 2)

	// Only verified contracts
	s.Require().EqualValues(3, contracts[0].Id)
	s.Require().True(contracts[0].Verified)
	s.Require().EqualValues(6, contracts[1].Id)
	s.Require().True(contracts[1].Verified)
}

func (s *StorageTestSuite) TestContractListWithTxFilterByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(1)
	contracts, err := s.storage.Contracts.ListWithTx(ctx, storage.ContractListFilter{
		Limit:      10,
		Offset:     0,
		Sort:       sdk.SortOrderAsc,
		SortField:  "id",
		IsVerified: false,
		TxId:       &txId,
	})
	s.Require().NoError(err)
	s.Require().Len(contracts, 1)
	s.Require().EqualValues(3, contracts[0].Id)
	s.Require().NotNil(contracts[0].TxId)
	s.Require().EqualValues(1, *contracts[0].TxId)
}

func (s *StorageTestSuite) TestContractListWithTxSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.ListWithTx(ctx, storage.ContractListFilter{
		Limit:      10,
		Offset:     0,
		Sort:       sdk.SortOrderDesc,
		SortField:  "id",
		IsVerified: false,
		TxId:       nil,
	})
	s.Require().NoError(err)
	s.Require().Len(contracts, 5)

	// Check descending order
	s.Require().EqualValues(7, contracts[0].Id)
	s.Require().EqualValues(6, contracts[1].Id)
	s.Require().EqualValues(3, contracts[4].Id)
}

func (s *StorageTestSuite) TestContractListWithTxSortByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.ListWithTx(ctx, storage.ContractListFilter{
		Limit:      10,
		Offset:     0,
		Sort:       sdk.SortOrderAsc,
		SortField:  "height",
		IsVerified: false,
		TxId:       nil,
	})
	s.Require().NoError(err)
	s.Require().Len(contracts, 5)

	// Check ascending order by height
	s.Require().EqualValues(100, contracts[0].Height)
	s.Require().EqualValues(200, contracts[1].Height)
	s.Require().EqualValues(300, contracts[2].Height)
	s.Require().EqualValues(400, contracts[3].Height)
	s.Require().EqualValues(500, contracts[4].Height)
}

func (s *StorageTestSuite) TestContractListWithTxLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.ListWithTx(ctx, storage.ContractListFilter{
		Limit:      2,
		Offset:     1,
		Sort:       sdk.SortOrderAsc,
		SortField:  "id",
		IsVerified: false,
		TxId:       nil,
	})
	s.Require().NoError(err)
	s.Require().Len(contracts, 2)

	// With offset=1, should skip first contract
	s.Require().EqualValues(4, contracts[0].Id)
	s.Require().EqualValues(5, contracts[1].Id)
}

func (s *StorageTestSuite) TestContractListWithTxEmptyResult() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.ListWithTx(ctx, storage.ContractListFilter{
		Limit:      10,
		Offset:     100,
		Sort:       sdk.SortOrderAsc,
		SortField:  "id",
		IsVerified: false,
		TxId:       nil,
	})
	s.Require().NoError(err)
	s.Require().Len(contracts, 0)
}

func (s *StorageTestSuite) TestContractPendingMetadata() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Use retry delay of 1 hour - should return contracts that haven't been updated in last hour
	contracts, err := s.storage.Contracts.PendingMetadata(ctx, 1*time.Hour, 10)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(contracts), 1)

	// Check that all returned contracts have pending status and metadata_link
	for _, contract := range contracts {
		s.Require().EqualValues(types.Pending, contract.Status)
		s.Require().NotEmpty(contract.MetadataLink)
		s.Require().NotNil(contract.Address)
	}
}

func (s *StorageTestSuite) TestContractPendingMetadataWithLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Test with limit 1
	contracts, err := s.storage.Contracts.PendingMetadata(ctx, 1*time.Hour, 1)
	s.Require().NoError(err)
	s.Require().LessOrEqual(len(contracts), 1)
}

func (s *StorageTestSuite) TestContractPendingMetadataRecentRetry() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Use very short retry delay - should exclude contracts updated recently
	// unless retry_count = 0
	contracts, err := s.storage.Contracts.PendingMetadata(ctx, 1*time.Nanosecond, 10)
	s.Require().NoError(err)

	// Should only return contracts with retry_count = 0 or updated long ago
	for _, contract := range contracts {
		s.Require().EqualValues(types.Pending, contract.Status)
		if contract.RetryCount > 0 {
			// If retry_count > 0, updated_at should be older than threshold
			threshold := time.Now().UTC().Add(-1 * time.Nanosecond)
			s.Require().True(contract.UpdatedAt.Before(threshold))
		}
	}
}

func (s *StorageTestSuite) TestContractPendingMetadataExcludesNonPending() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.PendingMetadata(ctx, 1*time.Hour, 10)
	s.Require().NoError(err)

	// Verify that no 'success' or 'failed' status contracts are returned
	for _, contract := range contracts {
		s.Require().NotEqual(types.Success, contract.Status)
		s.Require().NotEqual(types.Failed, contract.Status)
	}
}

func (s *StorageTestSuite) TestContractPendingMetadataExcludesEmptyMetadataLink() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	contracts, err := s.storage.Contracts.PendingMetadata(ctx, 1*time.Hour, 10)
	s.Require().NoError(err)

	// Verify that all contracts have non-empty metadata_link
	for _, contract := range contracts {
		s.Require().NotEmpty(contract.MetadataLink)
	}
}

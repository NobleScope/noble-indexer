package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestUserOpsFilterBasic() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 3)

	// Check first user op
	s.Require().EqualValues(1, userOps[0].Id)
	s.Require().EqualValues(1, userOps[0].Height)
	s.Require().EqualValues(1, userOps[0].TxId)
	s.Require().EqualValues(1, userOps[0].SenderId)
	s.Require().EqualValues(2, userOps[0].BundlerId)
	s.Require().NotNil(userOps[0].PaymasterId)
	s.Require().EqualValues(3, *userOps[0].PaymasterId)
	s.Require().True(userOps[0].Success)

	// Check that relations are loaded
	s.Require().NotNil(userOps[0].Tx.Hash)
	s.Require().EqualValues("0x90f5df4e03620cc55d3ea295bf8826f84465065340cb6d0d095166dd2465f283", userOps[0].Tx.Hash.Hex())
	s.Require().EqualValues("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2", userOps[0].Sender.Hash.Hex())
	s.Require().EqualValues("0xaa725ef35d90060a8cdfb77e324a9b770ca7e127", userOps[0].Bundler.Hash.Hex())
	s.Require().NotNil(userOps[0].Paymaster)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", userOps[0].Paymaster.Hash.Hex())
}

func (s *StorageTestSuite) TestUserOpsFilterDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 3)

	// Check descending order
	s.Require().EqualValues(3, userOps[0].Id)
	s.Require().EqualValues(3, userOps[0].Height)
	s.Require().EqualValues(1, userOps[2].Id)
	s.Require().EqualValues(1, userOps[2].Height)
}

func (s *StorageTestSuite) TestUserOpsFilterByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(2)
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: &height,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(2, userOps[0].Id)
	s.Require().EqualValues(2, userOps[0].Height)
	s.Require().False(userOps[0].Success)
}

func (s *StorageTestSuite) TestUserOpsFilterByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txId := uint64(4)
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		TxId:   &txId,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(2, userOps[0].Id)
	s.Require().EqualValues(4, userOps[0].TxId)
}

func (s *StorageTestSuite) TestUserOpsFilterByBundlerId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	bundlerId := uint64(3)
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:     10,
		Offset:    0,
		Sort:      sdk.SortOrderAsc,
		BundlerId: &bundlerId,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(2, userOps[0].Id)
	s.Require().EqualValues(3, userOps[0].BundlerId)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", userOps[0].Bundler.Hash.Hex())
}

func (s *StorageTestSuite) TestUserOpsFilterByPaymasterId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	paymasterId := uint64(3)
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:       10,
		Offset:      0,
		Sort:        sdk.SortOrderAsc,
		PaymasterId: &paymasterId,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(1, userOps[0].Id)
	s.Require().NotNil(userOps[0].PaymasterId)
	s.Require().EqualValues(3, *userOps[0].PaymasterId)
	s.Require().NotNil(userOps[0].Paymaster)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", userOps[0].Paymaster.Hash.Hex())
}

func (s *StorageTestSuite) TestUserOpsFilterBySuccessTrue() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	success := true
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:   10,
		Offset:  0,
		Sort:    sdk.SortOrderAsc,
		Success: &success,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 2)

	for _, op := range userOps {
		s.Require().True(op.Success)
	}

	s.Require().EqualValues(1, userOps[0].Id)
	s.Require().EqualValues(3, userOps[1].Id)
}

func (s *StorageTestSuite) TestUserOpsFilterBySuccessFalse() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	success := false
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:   10,
		Offset:  0,
		Sort:    sdk.SortOrderAsc,
		Success: &success,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(2, userOps[0].Id)
	s.Require().False(userOps[0].Success)
}

func (s *StorageTestSuite) TestUserOpsFilterByTimeRange() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:00Z")

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:    10,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(2, userOps[0].Id)
	s.Require().True(userOps[0].Time.Equal(timeFrom) || userOps[0].Time.After(timeFrom))
	s.Require().True(userOps[0].Time.Before(timeTo))
}

func (s *StorageTestSuite) TestUserOpsFilterTimeFromOnly() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:    10,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
		TimeFrom: timeFrom,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 2)

	for _, op := range userOps {
		s.Require().True(op.Time.Equal(timeFrom) || op.Time.After(timeFrom))
	}

	s.Require().EqualValues(2, userOps[0].Id)
	s.Require().EqualValues(3, userOps[1].Id)
}

func (s *StorageTestSuite) TestUserOpsFilterTimeToOnly() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	timeTo, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		TimeTo: timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(1, userOps[0].Id)
	s.Require().True(userOps[0].Time.Before(timeTo))
}

func (s *StorageTestSuite) TestUserOpsFilterLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  1,
		Offset: 1,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(2, userOps[0].Id)
}

func (s *StorageTestSuite) TestUserOpsFilterEmptyResult() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 100,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 0)
}

func (s *StorageTestSuite) TestUserOpsFilterNullPaymaster() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(2)
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: &height,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	// Record 2 has no paymaster
	s.Require().Nil(userOps[0].PaymasterId)
	s.Require().Nil(userOps[0].Paymaster)
}

func (s *StorageTestSuite) TestUserOpsFilterCombinedFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(1)
	success := true
	paymasterId := uint64(3)

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:       10,
		Offset:      0,
		Sort:        sdk.SortOrderAsc,
		Height:      &height,
		Success:     &success,
		PaymasterId: &paymasterId,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	op := userOps[0]
	s.Require().EqualValues(1, op.Id)
	s.Require().EqualValues(1, op.Height)
	s.Require().True(op.Success)
	s.Require().NotNil(op.PaymasterId)
	s.Require().EqualValues(3, *op.PaymasterId)
}

func (s *StorageTestSuite) TestUserOpsFilterNonExistentHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(999)
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		Height: &height,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 0)
}

func (s *StorageTestSuite) TestUserOpsFilterDescWithSuccess() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	success := true
	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:   10,
		Offset:  0,
		Sort:    sdk.SortOrderDesc,
		Success: &success,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 2)

	// Check descending order among successful ops
	s.Require().EqualValues(3, userOps[0].Id)
	s.Require().EqualValues(1, userOps[1].Id)

	// Verify JOIN fields are loaded
	s.Require().NotNil(userOps[0].Tx.Hash)
	s.Require().NotNil(userOps[0].Sender.Hash)
	s.Require().NotNil(userOps[0].Bundler.Hash)
}

func (s *StorageTestSuite) TestUserOpsFilterLimitExceedsTotal() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  100,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 3)

	s.Require().EqualValues(1, userOps[0].Id)
	s.Require().EqualValues(3, userOps[2].Id)
}

func (s *StorageTestSuite) TestUserOpsFilterMinimalLimit() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  1,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(1, userOps[0].Id)

	// Verify JOIN fields work with limit=1
	s.Require().NotNil(userOps[0].Tx.Hash)
	s.Require().NotNil(userOps[0].Sender.Hash)
	s.Require().NotNil(userOps[0].Bundler.Hash)
}

func (s *StorageTestSuite) TestUserOpsFilterExactBoundary() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:  10,
		Offset: 2,
		Sort:   sdk.SortOrderAsc,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(3, userOps[0].Id)
	s.Require().EqualValues(3, userOps[0].Height)
}

func (s *StorageTestSuite) TestUserOpsFilterHeightAndTimeRange() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(2)
	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-03T00:00:00Z")

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:    10,
		Offset:   0,
		Sort:     sdk.SortOrderAsc,
		Height:   &height,
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	s.Require().EqualValues(2, userOps[0].Id)
	s.Require().EqualValues(2, userOps[0].Height)
	s.Require().True(userOps[0].Time.Equal(timeFrom) || userOps[0].Time.After(timeFrom))
	s.Require().True(userOps[0].Time.Before(timeTo))
}

func (s *StorageTestSuite) TestUserOpsFilterAllFilters() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	height := uint64(1)
	txId := uint64(1)
	bundlerId := uint64(2)
	paymasterId := uint64(3)
	success := true
	timeFrom, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	timeTo, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")

	userOps, err := s.storage.ERC4337UserOps.Filter(ctx, storage.ERC4337UserOpsListFilter{
		Limit:       10,
		Offset:      0,
		Sort:        sdk.SortOrderAsc,
		Height:      &height,
		TxId:        &txId,
		BundlerId:   &bundlerId,
		PaymasterId: &paymasterId,
		Success:     &success,
		TimeFrom:    timeFrom,
		TimeTo:      timeTo,
	})
	s.Require().NoError(err)
	s.Require().Len(userOps, 1)

	op := userOps[0]
	s.Require().EqualValues(1, op.Id)
	s.Require().EqualValues(1, op.Height)
	s.Require().EqualValues(1, op.TxId)
	s.Require().EqualValues(2, op.BundlerId)
	s.Require().NotNil(op.PaymasterId)
	s.Require().EqualValues(3, *op.PaymasterId)
	s.Require().True(op.Success)
	s.Require().True(op.Time.Equal(timeFrom) || op.Time.After(timeFrom))
	s.Require().True(op.Time.Before(timeTo))
}

func (s *StorageTestSuite) TestUserOpsByHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("aabbccddee0011223344556677889900aabbccddee0011223344556677889900")
	op, err := s.storage.ERC4337UserOps.ByHash(ctx, hash)
	s.Require().NoError(err)

	s.Require().EqualValues(1, op.Id)
	s.Require().EqualValues(1, op.Height)
	s.Require().EqualValues(1, op.TxId)
	s.Require().EqualValues(1, op.SenderId)
	s.Require().EqualValues(2, op.BundlerId)
	s.Require().NotNil(op.PaymasterId)
	s.Require().EqualValues(3, *op.PaymasterId)
	s.Require().True(op.Success)

	// Check JOIN relations
	s.Require().EqualValues("0x90f5df4e03620cc55d3ea295bf8826f84465065340cb6d0d095166dd2465f283", op.Tx.Hash.Hex())
	s.Require().EqualValues("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2", op.Sender.Hash.Hex())
	s.Require().EqualValues("0xaa725ef35d90060a8cdfb77e324a9b770ca7e127", op.Bundler.Hash.Hex())
	s.Require().NotNil(op.Paymaster)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", op.Paymaster.Hash.Hex())
}

func (s *StorageTestSuite) TestUserOpsByHashNullPaymaster() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("1122334455667788990011223344556677889900aabbccddeeff001122334455")
	op, err := s.storage.ERC4337UserOps.ByHash(ctx, hash)
	s.Require().NoError(err)

	s.Require().EqualValues(2, op.Id)
	s.Require().False(op.Success)
	s.Require().Nil(op.PaymasterId)
	s.Require().Nil(op.Paymaster)

	// Other JOIN relations still loaded
	s.Require().NotNil(op.Tx.Hash)
	s.Require().NotNil(op.Sender.Hash)
	s.Require().NotNil(op.Bundler.Hash)
}

func (s *StorageTestSuite) TestUserOpsByHashNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash := pkgTypes.MustDecodeHex("0000000000000000000000000000000000000000000000000000000000000000")
	_, err := s.storage.ERC4337UserOps.ByHash(ctx, hash)
	s.Require().Error(err)
}

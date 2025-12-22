package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestByHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	address, err := s.storage.Addresses.ByHash(ctx, types.MustDecodeHex("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2"))
	s.Require().NoError(err)
	s.Require().EqualValues(1, address.Id)
	s.Require().EqualValues(500, address.FirstHeight)
	s.Require().EqualValues(550, address.LastHeight)
	s.Require().EqualValues("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2", address.Hash.Hex())
	s.Require().False(address.IsContract)
	s.Require().EqualValues(5, address.TxsCount)
	s.Require().EqualValues(0, address.ContractsCount)
	s.Require().EqualValues(10, address.Interactions)
	s.Require().NotNil(address.Balance)
	s.Require().EqualValues(1, address.Balance.Id)
	s.Require().EqualValues("noble", address.Balance.Currency)
	s.Require().EqualValues("1000000", address.Balance.Value.String())
}

func (s *StorageTestSuite) TestByHashWithoutBalance() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	address, err := s.storage.Addresses.ByHash(ctx, types.MustDecodeHex("0xaa725ef35d90060a8cdfb77e324a9b770ca7e127"))
	s.Require().NoError(err)
	s.Require().EqualValues(2, address.Id)
	s.Require().EqualValues(200, address.FirstHeight)
	s.Require().EqualValues(650, address.LastHeight)
	s.Require().EqualValues("0xaa725ef35d90060a8cdfb77e324a9b770ca7e127", address.Hash.Hex())
	s.Require().False(address.IsContract)
	s.Require().EqualValues(3, address.TxsCount)
	s.Require().EqualValues(0, address.ContractsCount)
	s.Require().EqualValues(5, address.Interactions)

	s.Require().Nil(address.Balance)
}

func (s *StorageTestSuite) TestByHashContract() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	address, err := s.storage.Addresses.ByHash(ctx, types.MustDecodeHex("0x30f055506ba543ea0942dc8ca03f596ab75bc879"))
	s.Require().NoError(err)
	s.Require().EqualValues(3, address.Id)
	s.Require().EqualValues(100, address.FirstHeight)
	s.Require().EqualValues(300, address.LastHeight)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", address.Hash.Hex())
	s.Require().True(address.IsContract)
	s.Require().EqualValues(10, address.TxsCount)
	s.Require().EqualValues(2, address.ContractsCount)
	s.Require().EqualValues(25, address.Interactions)
	s.Require().NotNil(address.Balance)
	s.Require().EqualValues(3, address.Balance.Id)
	s.Require().EqualValues("noble", address.Balance.Currency)
	s.Require().EqualValues("5000000", address.Balance.Value.String())
}

func (s *StorageTestSuite) TestByHashNotFound() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.storage.Addresses.ByHash(ctx, types.MustDecodeHex("0x0000000000000000000000000000000000000001"))
	s.Require().Error(err)
	s.Require().ErrorIs(err, sql.ErrNoRows)
}

func (s *StorageTestSuite) TestListWithBalance() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         10,
		Offset:        0,
		Sort:          sdk.SortOrderAsc,
		SortField:     "id",
		OnlyContracts: false,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 3)

	// Address 1
	s.Require().EqualValues(1, addresses[0].Id)
	s.Require().EqualValues(500, addresses[0].FirstHeight)
	s.Require().EqualValues(550, addresses[0].LastHeight)
	s.Require().EqualValues("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2", addresses[0].Hash.Hex())
	s.Require().False(addresses[0].IsContract)
	s.Require().EqualValues(5, addresses[0].TxsCount)
	s.Require().EqualValues(0, addresses[0].ContractsCount)
	s.Require().EqualValues(10, addresses[0].Interactions)
	s.Require().NotNil(addresses[0].Balance)
	s.Require().EqualValues(1, addresses[0].Balance.Id)
	s.Require().EqualValues("noble", addresses[0].Balance.Currency)
	s.Require().EqualValues("1000000", addresses[0].Balance.Value.String())

	// Address 2
	s.Require().EqualValues(2, addresses[1].Id)
	s.Require().EqualValues(200, addresses[1].FirstHeight)
	s.Require().EqualValues(650, addresses[1].LastHeight)
	s.Require().EqualValues("0xaa725ef35d90060a8cdfb77e324a9b770ca7e127", addresses[1].Hash.Hex())
	s.Require().False(addresses[1].IsContract)
	s.Require().EqualValues(3, addresses[1].TxsCount)
	s.Require().EqualValues(0, addresses[1].ContractsCount)
	s.Require().EqualValues(5, addresses[1].Interactions)
	s.Require().Nil(addresses[1].Balance)

	// Address 3
	s.Require().EqualValues(3, addresses[2].Id)
	s.Require().EqualValues(100, addresses[2].FirstHeight)
	s.Require().EqualValues(300, addresses[2].LastHeight)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", addresses[2].Hash.Hex())
	s.Require().True(addresses[2].IsContract)
	s.Require().EqualValues(10, addresses[2].TxsCount)
	s.Require().EqualValues(2, addresses[2].ContractsCount)
	s.Require().EqualValues(25, addresses[2].Interactions)
	s.Require().NotNil(addresses[2].Balance)
	s.Require().EqualValues(3, addresses[2].Balance.Id)
	s.Require().EqualValues("noble", addresses[2].Balance.Currency)
	s.Require().EqualValues("5000000", addresses[2].Balance.Value.String())
}

func (s *StorageTestSuite) TestListWithBalanceOnlyContracts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         10,
		Offset:        0,
		Sort:          sdk.SortOrderAsc,
		SortField:     "id",
		OnlyContracts: true,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 1)

	// Only contract address should be returned
	s.Require().EqualValues(3, addresses[0].Id)
	s.Require().EqualValues(100, addresses[0].FirstHeight)
	s.Require().EqualValues(300, addresses[0].LastHeight)
	s.Require().EqualValues("0x30f055506ba543ea0942dc8ca03f596ab75bc879", addresses[0].Hash.Hex())
	s.Require().True(addresses[0].IsContract)
	s.Require().EqualValues(10, addresses[0].TxsCount)
	s.Require().EqualValues(2, addresses[0].ContractsCount)
	s.Require().EqualValues(25, addresses[0].Interactions)
	s.Require().NotNil(addresses[0].Balance)
	s.Require().EqualValues("noble", addresses[0].Balance.Currency)
	s.Require().EqualValues("5000000", addresses[0].Balance.Value.String())
}

func (s *StorageTestSuite) TestListWithBalanceLimitOffset() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         1,
		Offset:        1,
		Sort:          sdk.SortOrderAsc,
		SortField:     "id",
		OnlyContracts: false,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 1)

	// With offset=1, should return the second address (id=2)
	s.Require().EqualValues(2, addresses[0].Id)
	s.Require().EqualValues(200, addresses[0].FirstHeight)
	s.Require().EqualValues(650, addresses[0].LastHeight)
	s.Require().EqualValues("0xaa725ef35d90060a8cdfb77e324a9b770ca7e127", addresses[0].Hash.Hex())
	s.Require().False(addresses[0].IsContract)
	s.Require().Nil(addresses[0].Balance)
}

func (s *StorageTestSuite) TestListWithBalanceSortDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         10,
		Offset:        0,
		Sort:          sdk.SortOrderDesc,
		SortField:     "id",
		OnlyContracts: false,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 3)

	// Check descending order by ID
	s.Require().EqualValues(3, addresses[0].Id)
	s.Require().EqualValues(100, addresses[0].FirstHeight)
	s.Require().EqualValues(2, addresses[1].Id)
	s.Require().EqualValues(200, addresses[1].FirstHeight)
	s.Require().EqualValues(1, addresses[2].Id)
	s.Require().EqualValues(500, addresses[2].FirstHeight)
}

func (s *StorageTestSuite) TestListWithBalanceSortByFirstHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         10,
		Offset:        0,
		Sort:          sdk.SortOrderAsc,
		SortField:     "first_height",
		OnlyContracts: false,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 3)

	// Check sorting by first_height ascending
	s.Require().EqualValues(3, addresses[0].Id)
	s.Require().EqualValues(100, addresses[0].FirstHeight)
	s.Require().EqualValues(2, addresses[1].Id)
	s.Require().EqualValues(200, addresses[1].FirstHeight)
	s.Require().EqualValues(1, addresses[2].Id)
	s.Require().EqualValues(500, addresses[2].FirstHeight)
}

func (s *StorageTestSuite) TestListWithBalanceSortByLastHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         10,
		Offset:        0,
		Sort:          sdk.SortOrderDesc,
		SortField:     "last_height",
		OnlyContracts: false,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 3)

	// Check sorting by last_height descending
	s.Require().EqualValues(2, addresses[0].Id)
	s.Require().EqualValues(650, addresses[0].LastHeight)
	s.Require().EqualValues(1, addresses[1].Id)
	s.Require().EqualValues(550, addresses[1].LastHeight)
	s.Require().EqualValues(3, addresses[2].Id)
	s.Require().EqualValues(300, addresses[2].LastHeight)
}

func (s *StorageTestSuite) TestListWithBalanceSortByValue() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         10,
		Offset:        0,
		Sort:          sdk.SortOrderAsc,
		SortField:     "value",
		OnlyContracts: false,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 2)
	s.Require().EqualValues(1, addresses[0].Id)
	s.Require().NotNil(addresses[0].Balance)
	s.Require().EqualValues("1000000", addresses[0].Balance.Value.String())

	s.Require().EqualValues(3, addresses[1].Id)
	s.Require().NotNil(addresses[1].Balance)
	s.Require().EqualValues("5000000", addresses[1].Balance.Value.String())
}

func (s *StorageTestSuite) TestListWithBalanceSortByValueDesc() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	addresses, err := s.storage.Addresses.ListWithBalance(ctx, storage.AddressListFilter{
		Limit:         10,
		Offset:        0,
		Sort:          sdk.SortOrderDesc,
		SortField:     "value",
		OnlyContracts: false,
	})
	s.Require().NoError(err)
	s.Require().Len(addresses, 2)

	s.Require().EqualValues(3, addresses[0].Id)
	s.Require().NotNil(addresses[0].Balance)
	s.Require().EqualValues("5000000", addresses[0].Balance.Value.String())

	s.Require().EqualValues(1, addresses[1].Id)
	s.Require().NotNil(addresses[1].Balance)
	s.Require().EqualValues("1000000", addresses[1].Balance.Value.String())
}

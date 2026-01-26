package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

func uint64Ptr(v uint64) *uint64 {
	return &v
}

// TransactionTestSuite -
type TransactionTestSuite struct {
	suite.Suite
	psqlContainer *database.PostgreSQLContainer
	storage       Storage
}

// SetupSuite -
func (s *TransactionTestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer ctxCancel()

	psqlContainer, err := database.NewPostgreSQLContainer(ctx, database.PostgreSQLContainerConfig{
		User:     "user",
		Password: "password",
		Database: "db_test",
		Port:     5432,
		Image:    "timescale/timescaledb-ha:pg17.6-ts2.22.1-all",
	})
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	strg, err := Create(ctx, config.Database{
		Kind:     config.DBKindPostgres,
		User:     s.psqlContainer.Config.User,
		Database: s.psqlContainer.Config.Database,
		Password: s.psqlContainer.Config.Password,
		Host:     s.psqlContainer.Config.Host,
		Port:     s.psqlContainer.MappedPort().Int(),
	}, "../../../database", false)
	s.Require().NoError(err)
	s.storage = strg
}

// TearDownSuite -
func (s *TransactionTestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.storage.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

func (s *TransactionTestSuite) BeforeTest(suiteName, testName string) {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("timescaledb"),
		testfixtures.Directory("../../../test/data"),
		testfixtures.UseAlterConstraint(),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())
}

func TestSuiteTransaction_Run(t *testing.T) {
	suite.Run(t, new(TransactionTestSuite))
}

func (s *TransactionTestSuite) TestSaveAddresses() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	addresses := []*storage.Address{
		{
			Hash:           pkgTypes.MustDecodeHex("0x1111111111111111111111111111111111111111"),
			FirstHeight:    1000,
			LastHeight:     1000,
			IsContract:     false,
			TxsCount:       1,
			ContractsCount: 0,
			Interactions:   1,
		},
		{
			Hash:           pkgTypes.MustDecodeHex("0x2222222222222222222222222222222222222222"),
			FirstHeight:    1001,
			LastHeight:     1001,
			IsContract:     true,
			TxsCount:       1,
			ContractsCount: 1,
			Interactions:   2,
		},
	}

	count, err := tx.SaveAddresses(ctx, addresses...)
	s.Require().NoError(err)
	s.Require().EqualValues(2, count)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	s.Require().Greater(addresses[0].Id, uint64(0))
	s.Require().Greater(addresses[1].Id, uint64(0))
}

func (s *TransactionTestSuite) TestSaveAddressesUpdate() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	// Update existing address (id: 1)
	addresses := []*storage.Address{
		{
			Hash:           pkgTypes.MustDecodeHex("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2"),
			FirstHeight:    500,
			LastHeight:     1000,
			IsContract:     false,
			TxsCount:       3,
			ContractsCount: 0,
			Interactions:   5,
		},
	}

	count, err := tx.SaveAddresses(ctx, addresses...)
	s.Require().NoError(err)
	s.Require().EqualValues(0, count) // No new addresses

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify update
	addr, err := s.storage.Addresses.ByHash(ctx, pkgTypes.MustDecodeHex("0xa63d581a7fdab643c09f0524904b046cdb9ad9d2"))
	s.Require().NoError(err)
	s.Require().EqualValues(1000, addr.LastHeight)
	s.Require().EqualValues(8, addr.TxsCount)      // 5 + 3
	s.Require().EqualValues(15, addr.Interactions) // 10 + 5
}

func (s *TransactionTestSuite) TestSaveBalances() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	balances := []*storage.Balance{
		{
			Id:    1,
			Value: decimal.RequireFromString("500000"),
		},
		{
			Id:    2,
			Value: decimal.RequireFromString("1000000"),
		},
	}

	err = tx.SaveBalances(ctx, balances...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify balances were saved (SaveBalances adds value to existing balance)
	// Fixture: id=1 has value=1000000, id=3 has value=5000000
	var balance1 storage.Balance
	err = s.storage.Connection().DB().NewSelect().Model(&balance1).Where("id = ?", 1).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Equal(decimal.RequireFromString("1500000"), balance1.Value) // 1000000 + 500000

	var balance2 storage.Balance
	err = s.storage.Connection().DB().NewSelect().Model(&balance2).Where("id = ?", 2).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Equal(decimal.RequireFromString("1000000"), balance2.Value) // new balance
}

func (s *TransactionTestSuite) TestSaveContracts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	contracts := []*storage.Contract{
		{
			Id:     100,
			Height: 1000,
			Code:   pkgTypes.MustDecodeHex("0x6080604052"),
		},
	}

	_, err = tx.SaveContracts(ctx, contracts...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify contract was saved
	contract, err := s.storage.Contracts.GetByID(ctx, 100)
	s.Require().NoError(err)
	s.Require().EqualValues(1000, contract.Height)
	s.Require().Equal(pkgTypes.Hex(pkgTypes.MustDecodeHex("0x6080604052")), contract.Code)
}

func (s *TransactionTestSuite) TestSaveContractsUpdateMetadata() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	// Update existing contract (id: 4) with metadata
	abi := json.RawMessage(`[{"type":"function","name":"transfer"}]`)
	contracts := []*storage.Contract{
		{
			Id:              4,
			Height:          200,
			Verified:        true,
			ABI:             abi,
			CompilerVersion: "v0.8.20",
			Language:        "Solidity",
			Status:          types.Success,
		},
	}

	_, err = tx.SaveContracts(ctx, contracts...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify update
	contract, err := s.storage.Contracts.GetByID(ctx, 4)
	s.Require().NoError(err)
	s.Require().True(contract.Verified)
	s.Require().Equal("v0.8.20", contract.CompilerVersion)
	s.Require().Equal("Solidity", contract.Language)
}

func (s *TransactionTestSuite) TestSaveContractsNoOverwriteMetadata() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	// Try to update contract (id: 3) with empty metadata - should not overwrite
	contracts := []*storage.Contract{
		{
			Id:              3,
			Height:          100,
			Verified:        false, // false should not overwrite true
			CompilerVersion: "",    // empty should not overwrite existing
			Language:        "",    // empty should not overwrite existing
		},
	}

	_, err = tx.SaveContracts(ctx, contracts...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify that original values are preserved
	contract, err := s.storage.Contracts.GetByID(ctx, 3)
	s.Require().NoError(err)
	s.Require().True(contract.Verified)                                    // Should still be true
	s.Require().Equal("v0.8.20+commit.a1b79de6", contract.CompilerVersion) // Should be preserved
	s.Require().Equal("Solidity", contract.Language)                       // Should be preserved
}

func (s *TransactionTestSuite) TestSaveTokens() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	tokens := []*storage.Token{
		{
			TokenID:        decimal.NewFromInt(0),
			ContractId:     3,
			Type:           types.ERC20,
			Height:         100,
			LastHeight:     500,
			TransfersCount: 10,
			Supply:         decimal.RequireFromString("1000"),
			Status:         types.Pending,
		},
	}

	_, err = tx.SaveTokens(ctx, tokens...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify update - transfers_count and supply should be added
	token, err := s.storage.Token.GetByID(ctx, 1)
	s.Require().NoError(err)
	s.Require().EqualValues(500, token.LastHeight)
	s.Require().EqualValues(110, token.TransfersCount) // 100 + 10
}

func (s *TransactionTestSuite) TestSaveTokenMetadata() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	metadata := []byte(`{"name":"Updated Token"}`)
	tokens := []*storage.Token{
		{
			TokenID:      decimal.NewFromInt(0),
			ContractId:   4,
			Name:         "Updated Token",
			Symbol:       "UPD",
			Decimals:     18,
			MetadataLink: "https://example.com/updated",
			Metadata:     metadata,
			Status:       types.Success,
		},
	}

	err = tx.SaveTokenMetadata(ctx, tokens...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify update
	token, err := s.storage.Token.GetByID(ctx, 3)
	s.Require().NoError(err)
	s.Require().Equal("Updated Token", token.Name)
	s.Require().Equal("UPD", token.Symbol)
	s.Require().EqualValues(18, token.Decimals)
}

func (s *TransactionTestSuite) TestSaveTokenMetadataNoOverwrite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	// Try to update with empty values - should not overwrite
	tokens := []*storage.Token{
		{
			TokenID:    decimal.NewFromInt(0),
			ContractId: 3,
			Name:       "", // empty should not overwrite
			Symbol:     "", // empty should not overwrite
			Decimals:   0,  // 0 should not overwrite
			Status:     types.Pending,
		},
	}

	err = tx.SaveTokenMetadata(ctx, tokens...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify original values preserved
	token, err := s.storage.Token.GetByID(ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal("Test Token", token.Name)
	s.Require().Equal("TST", token.Symbol)
	s.Require().EqualValues(18, token.Decimals)
}

func (s *TransactionTestSuite) TestSaveTokenBalances() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	tokenBalances := []*storage.TokenBalance{
		{
			TokenID:    decimal.NewFromInt(0),
			ContractID: 3,
			AddressID:  1,
			Balance:    decimal.RequireFromString("500000000000000000"),
		},
	}

	result, err := tx.SaveTokenBalances(ctx, tokenBalances...)
	s.Require().NoError(err)
	s.Require().Len(result, 1)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify balance was added
	s.Require().Equal(decimal.RequireFromString("1500000000000000000"), result[0].Balance) // 1000000000000000000 + 500000000000000000
}

func (s *TransactionTestSuite) TestSaveProxyContracts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	proxyContracts := []*storage.ProxyContract{
		{
			Id:               100,
			ImplementationID: uint64Ptr(3),
			Height:           1000,
			Type:             types.EIP1967,
			Status:           types.Resolved,
		},
	}

	err = tx.SaveProxyContracts(ctx, proxyContracts...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify proxy contract was saved
	proxy, err := s.storage.ProxyContracts.GetByID(ctx, 100)
	s.Require().NoError(err)
	s.Require().EqualValues(1000, proxy.Height)
	s.Require().Equal(types.EIP1967, proxy.Type)
	s.Require().Equal(types.Resolved, proxy.Status)
	s.Require().NotNil(proxy.ImplementationID)
	s.Require().EqualValues(3, *proxy.ImplementationID)
}

func (s *TransactionTestSuite) TestRollbackBlock() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify block exists before rollback
	block, err := s.storage.Blocks.ByHeight(ctx, 100, false)
	s.Require().NoError(err)
	s.Require().EqualValues(100, block.Height)
	s.Require().EqualValues(pkgTypes.MustDecodeHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"), block.Hash)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.RollbackBlock(ctx, 100)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify block was deleted
	_, err = s.storage.Blocks.ByHeight(ctx, 100, false)
	s.Require().Error(err)
}

func (s *TransactionTestSuite) TestRollbackTxs() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify txs exist before rollback
	txsBefore, err := s.storage.Tx.ByHeight(ctx, 100, 100, 0, "asc")
	s.Require().NoError(err)
	s.Require().NotEmpty(txsBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	txs, err := tx.RollbackTxs(ctx, 100)
	s.Require().NoError(err)
	s.Require().NotNil(txs)
	s.Require().Len(txs, len(txsBefore))

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify txs were deleted
	txsAfter, err := s.storage.Tx.ByHeight(ctx, 100, 100, 0, "asc")
	s.Require().NoError(err)
	s.Require().Empty(txsAfter)
}

func (s *TransactionTestSuite) TestRollbackLogs() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify logs exist before rollback
	height := uint64(100)
	logsBefore, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{Height: &height, Limit: 100})
	s.Require().NoError(err)
	s.Require().NotEmpty(logsBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.RollbackLogs(ctx, 100)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify logs were deleted
	logsAfter, err := s.storage.Logs.Filter(ctx, storage.LogListFilter{Height: &height, Limit: 100})
	s.Require().NoError(err)
	s.Require().Empty(logsAfter)
}

func (s *TransactionTestSuite) TestRollbackTraces() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify traces exist before rollback
	height := uint64(100)
	tracesBefore, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{Height: &height, Limit: 100})
	s.Require().NoError(err)
	s.Require().NotEmpty(tracesBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	traces, err := tx.RollbackTraces(ctx, 100)
	s.Require().NoError(err)
	s.Require().NotNil(traces)
	s.Require().Len(traces, len(tracesBefore))

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify traces were deleted
	tracesAfter, err := s.storage.Trace.Filter(ctx, storage.TraceListFilter{Height: &height, Limit: 100})
	s.Require().NoError(err)
	s.Require().Empty(tracesAfter)
}

func (s *TransactionTestSuite) TestRollbackTransfers() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify transfers exist before rollback
	height := uint64(100)
	transfersBefore, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{Height: &height, Limit: 100})
	s.Require().NoError(err)
	s.Require().NotEmpty(transfersBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	transfers, err := tx.RollbackTransfers(ctx, 100)
	s.Require().NoError(err)
	s.Require().NotNil(transfers)
	s.Require().Len(transfers, len(transfersBefore))

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify transfers were deleted
	transfersAfter, err := s.storage.Transfer.Filter(ctx, storage.TransferListFilter{Height: &height, Limit: 100})
	s.Require().NoError(err)
	s.Require().Empty(transfersAfter)
}

func (s *TransactionTestSuite) TestRollbackTokens() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify tokens exist before rollback
	var tokensBefore []storage.Token
	err := s.storage.Connection().DB().NewSelect().Model(&tokensBefore).Where("height = ?", 100).Scan(ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(tokensBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	tokens, err := tx.RollbackTokens(ctx, 100)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)
	s.Require().Len(tokens, len(tokensBefore))

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify tokens were deleted
	var tokensAfter []storage.Token
	err = s.storage.Connection().DB().NewSelect().Model(&tokensAfter).Where("height = ?", 100).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Empty(tokensAfter)
}

func (s *TransactionTestSuite) TestRollbackContracts() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify contracts exist before rollback
	var contractsBefore []storage.Contract
	err := s.storage.Connection().DB().NewSelect().Model(&contractsBefore).Where("height = ?", 100).Scan(ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(contractsBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.RollbackContracts(ctx, 100)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify contracts were deleted
	var contractsAfter []storage.Contract
	err = s.storage.Connection().DB().NewSelect().Model(&contractsAfter).Where("height = ?", 100).Scan(ctx)
	s.Require().NoError(err)
	s.Require().Empty(contractsAfter)
}

func (s *TransactionTestSuite) TestDeleteBalances() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify balances exist before delete
	var balance1Before storage.Balance
	err := s.storage.Connection().DB().NewSelect().Model(&balance1Before).Where("id = ?", 1).Scan(ctx)
	s.Require().NoError(err)

	var balance3Before storage.Balance
	err = s.storage.Connection().DB().NewSelect().Model(&balance3Before).Where("id = ?", 3).Scan(ctx)
	s.Require().NoError(err)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.DeleteBalances(ctx, []uint64{1, 3})
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify balances were deleted
	var balance1After storage.Balance
	err = s.storage.Connection().DB().NewSelect().Model(&balance1After).Where("id = ?", 1).Scan(ctx)
	s.Require().Error(err)

	var balance3After storage.Balance
	err = s.storage.Connection().DB().NewSelect().Model(&balance3After).Where("id = ?", 3).Scan(ctx)
	s.Require().Error(err)
}

func (s *TransactionTestSuite) TestDeleteBalancesEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.DeleteBalances(ctx, []uint64{})
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestDeleteTokenBalances() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify token balances exist before delete (token_id=0, contract_id=3)
	tokenId := decimal.NewFromInt(0)
	contractId := uint64(3)
	balancesBefore, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		TokenId:    &tokenId,
		ContractId: &contractId,
		Limit:      100,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(balancesBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.DeleteTokenBalances(ctx, []string{"0"}, []uint64{3}, nil)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify token balances were deleted
	balancesAfter, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		TokenId:    &tokenId,
		ContractId: &contractId,
		Limit:      100,
	})
	s.Require().NoError(err)
	s.Require().Empty(balancesAfter)
}

func (s *TransactionTestSuite) TestDeleteTokenBalancesWithZeroBalances() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// Verify token balance exists before delete (token_id=1, contract_id=3, address_id=1)
	tokenId := decimal.NewFromInt(1)
	contractId := uint64(3)
	addressId := uint64(1)
	balancesBefore, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		TokenId:    &tokenId,
		ContractId: &contractId,
		AddressId:  &addressId,
		Limit:      100,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(balancesBefore)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	zeroBalances := []*storage.TokenBalance{
		{
			TokenID:    decimal.NewFromInt(1),
			ContractID: 3,
			AddressID:  1,
		},
	}

	err = tx.DeleteTokenBalances(ctx, []string{}, []uint64{}, zeroBalances)
	s.Require().NoError(err)

	s.Require().NoError(tx.Flush(ctx))
	s.Require().NoError(tx.Close(ctx))

	// Verify token balance was deleted
	balancesAfter, err := s.storage.TokenBalance.Filter(ctx, storage.TokenBalanceListFilter{
		TokenId:    &tokenId,
		ContractId: &contractId,
		AddressId:  &addressId,
		Limit:      100,
	})
	s.Require().NoError(err)
	s.Require().Empty(balancesAfter)
}

func (s *TransactionTestSuite) TestDeleteTokenBalancesMismatchedLengths() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.DeleteTokenBalances(ctx, []string{"0", "1"}, []uint64{3}, nil)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "tokenIds and contractIds must have same length")

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestState() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	// First create a state
	_, err := s.storage.Connection().DB().ExecContext(ctx, `
		INSERT INTO state (name, last_height, last_hash, last_time, total_tx, total_accounts, chain_id)
		VALUES ('indexer', 1000, '0x123', NOW(), 100, 50, 1)
		ON CONFLICT (name) DO UPDATE SET last_height = 1000
	`)
	s.Require().NoError(err)

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	state, err := tx.State(ctx, "indexer")
	s.Require().NoError(err)
	s.Require().EqualValues(1000, state.LastHeight)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestLastBlock() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	block, err := tx.LastBlock(ctx)
	s.Require().NoError(err)
	s.Require().Greater(block.Id, uint64(0))

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveTransactionsEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveTransactions(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveLogsEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveLogs(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveSourcesEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveSources(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveTracesEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveTraces(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveTransfersEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveTransfers(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveAddressesEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	count, err := tx.SaveAddresses(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(0, count)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveBalancesEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveBalances(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveContractsEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	_, err = tx.SaveContracts(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveTokensEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	_, err = tx.SaveTokens(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveTokenMetadataEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveTokenMetadata(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveTokenBalancesEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	result, err := tx.SaveTokenBalances(ctx)
	s.Require().NoError(err)
	s.Require().Nil(result)

	s.Require().NoError(tx.Close(ctx))
}

func (s *TransactionTestSuite) TestSaveProxyContractsEmpty() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	tx, err := BeginTransaction(ctx, s.storage.Transactable)
	s.Require().NoError(err)

	err = tx.SaveProxyContracts(ctx)
	s.Require().NoError(err)

	s.Require().NoError(tx.Close(ctx))
}

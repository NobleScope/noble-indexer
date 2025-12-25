package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/mock"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var (
	testToken1 = storage.Token{
		Id:             1,
		TokenID:        decimal.NewFromInt(0),
		ContractId:     1,
		Type:           types.ERC20,
		Height:         100,
		LastHeight:     200,
		Name:           "Test Token",
		Symbol:         "TST",
		Decimals:       18,
		TransfersCount: 100,
		Supply:         decimal.NewFromInt(1000000),
		MetadataLink:   "https://example.com/metadata/0",
		Status:         types.Success,
	}

	testToken2 = storage.Token{
		Id:             2,
		TokenID:        decimal.NewFromInt(1),
		ContractId:     1,
		Type:           types.ERC721,
		Height:         101,
		LastHeight:     201,
		Name:           "NFT Token",
		Symbol:         "NFT",
		Decimals:       0,
		TransfersCount: 50,
		Supply:         decimal.NewFromInt(1),
		MetadataLink:   "https://example.com/metadata/1",
		Status:         types.Success,
	}

	testToken3 = storage.Token{
		Id:             3,
		TokenID:        decimal.NewFromInt(2),
		ContractId:     2,
		Type:           types.ERC1155,
		Height:         102,
		LastHeight:     202,
		Name:           "Multi Token",
		Symbol:         "MLT",
		Decimals:       0,
		TransfersCount: 25,
		Supply:         decimal.NewFromInt(100),
		MetadataLink:   "https://example.com/metadata/2",
		Status:         types.Pending,
	}

	testTransfer1 = storage.Transfer{
		Id:            1,
		Height:        100,
		Time:          testTime,
		TokenID:       decimal.NewFromInt(0),
		Amount:        decimal.NewFromInt(1000),
		Type:          types.Transfer,
		ContractId:    1,
		FromAddressId: uint64Ptr(1),
		ToAddressId:   uint64Ptr(2),
		TxID:          1,
		FromAddress:   &testFromAddress,
		ToAddress:     &testToAddress,
	}

	testTransfer2 = storage.Transfer{
		Id:            2,
		Height:        101,
		Time:          testTime.Add(time.Hour),
		TokenID:       decimal.NewFromInt(1),
		Amount:        decimal.NewFromInt(1),
		Type:          types.Mint,
		ContractId:    1,
		FromAddressId: nil,
		ToAddressId:   uint64Ptr(2),
		TxID:          2,
		FromAddress:   nil,
		ToAddress:     &testToAddress,
	}

	testTransfer3 = storage.Transfer{
		Id:            3,
		Height:        102,
		Time:          testTime.Add(2 * time.Hour),
		TokenID:       decimal.NewFromInt(2),
		Amount:        decimal.NewFromInt(500),
		Type:          types.Burn,
		ContractId:    2,
		FromAddressId: uint64Ptr(1),
		ToAddressId:   nil,
		TxID:          3,
		FromAddress:   &testFromAddress,
		ToAddress:     nil,
	}

	testTokenBalance1 = storage.TokenBalance{
		Id:         1,
		TokenID:    decimal.NewFromInt(0),
		ContractID: 1,
		AddressID:  1,
		Balance:    decimal.NewFromInt(5000),
		Address:    testFromAddress,
	}

	testTokenBalance2 = storage.TokenBalance{
		Id:         2,
		TokenID:    decimal.NewFromInt(1),
		ContractID: 1,
		AddressID:  2,
		Balance:    decimal.NewFromInt(1),
		Address:    testToAddress,
	}

	testTokenBalance3 = storage.TokenBalance{
		Id:         3,
		TokenID:    decimal.NewFromInt(2),
		ContractID: 2,
		AddressID:  1,
		Balance:    decimal.NewFromInt(100),
		Address:    testFromAddress,
	}
)

// TokenHandlerTestSuite -
type TokenHandlerTestSuite struct {
	suite.Suite
	token    *mock.MockIToken
	transfer *mock.MockITransfer
	tbs      *mock.MockITokenBalance
	address  *mock.MockIAddress
	tx       *mock.MockITx
	echo     *echo.Echo
	handler  *TokenHandler
	ctrl     *gomock.Controller
}

// SetupSuite -
func (s *TokenHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.token = mock.NewMockIToken(s.ctrl)
	s.transfer = mock.NewMockITransfer(s.ctrl)
	s.tbs = mock.NewMockITokenBalance(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.tx = mock.NewMockITx(s.ctrl)
	s.handler = NewTokenHandler(s.token, s.transfer, s.tbs, s.address, s.tx)
}

// TearDownSuite -
func (s *TokenHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteTokenHandler_Run(t *testing.T) {
	suite.Run(t, new(TokenHandlerTestSuite))
}

// ====================================
// Token List Tests
// ====================================

// TestListSuccess tests successful retrieval of tokens with default parameters
func (s *TokenHandlerTestSuite) TestListSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.token.EXPECT().
		Filter(gomock.Any(), storage.TokenListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TokenType{},
		}).
		Return([]storage.Token{testToken1, testToken2, testToken3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tokens []responses.Token
	err := json.NewDecoder(rec.Body).Decode(&tokens)
	s.Require().NoError(err)
	s.Require().Len(tokens, 3)

	s.Require().EqualValues(1, tokens[0].Id)
	s.Require().Equal("Test Token", tokens[0].Name)
	s.Require().Equal("TST", tokens[0].Symbol)
	s.Require().EqualValues(18, tokens[0].Decimals)
}

// TestListWithLimit tests list with custom limit parameter
func (s *TokenHandlerTestSuite) TestListWithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.token.EXPECT().
		Filter(gomock.Any(), storage.TokenListFilter{
			Limit:  5,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TokenType{},
		}).
		Return([]storage.Token{testToken1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tokens []responses.Token
	err := json.NewDecoder(rec.Body).Decode(&tokens)
	s.Require().NoError(err)
	s.Require().Len(tokens, 1)
}

// TestListWithOffset tests list with offset parameter
func (s *TokenHandlerTestSuite) TestListWithOffset() {
	q := make(url.Values)
	q.Set("offset", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.token.EXPECT().
		Filter(gomock.Any(), storage.TokenListFilter{
			Limit:  10,
			Offset: 1,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TokenType{},
		}).
		Return([]storage.Token{testToken2, testToken3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tokens []responses.Token
	err := json.NewDecoder(rec.Body).Decode(&tokens)
	s.Require().NoError(err)
	s.Require().Len(tokens, 2)
}

// TestListAscOrder tests list with ascending sort order
func (s *TokenHandlerTestSuite) TestListAscOrder() {
	q := make(url.Values)
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.token.EXPECT().
		Filter(gomock.Any(), storage.TokenListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
			Type:   []types.TokenType{},
		}).
		Return([]storage.Token{testToken1, testToken2, testToken3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestListWithContract tests filtering tokens by contract
func (s *TokenHandlerTestSuite) TestListWithContract() {
	q := make(url.Values)
	q.Set("contract", testAddressHex1.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	contractId := uint64(1)
	s.token.EXPECT().
		Filter(gomock.Any(), storage.TokenListFilter{
			Limit:      10,
			Offset:     0,
			Sort:       sdk.SortOrderDesc,
			Type:       []types.TokenType{},
			ContractId: &contractId,
		}).
		Return([]storage.Token{testToken1, testToken2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tokens []responses.Token
	err := json.NewDecoder(rec.Body).Decode(&tokens)
	s.Require().NoError(err)
	s.Require().Len(tokens, 2)
}

// TestListWithType tests filtering tokens by type
func (s *TokenHandlerTestSuite) TestListWithType() {
	q := make(url.Values)
	q.Set("type", "ERC20,ERC721")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.token.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TokenListFilter) ([]storage.Token, error) {
			s.Require().Equal(10, filter.Limit)
			s.Require().Equal(0, filter.Offset)
			s.Require().Equal(sdk.SortOrderDesc, filter.Sort)
			s.Require().Len(filter.Type, 2)
			return []storage.Token{testToken1, testToken2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tokens []responses.Token
	err := json.NewDecoder(rec.Body).Decode(&tokens)
	s.Require().NoError(err)
	s.Require().Len(tokens, 2)
}

// TestListEmptyResult tests list when no tokens are found
func (s *TokenHandlerTestSuite) TestListEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.token.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		Return([]storage.Token{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tokens []responses.Token
	err := json.NewDecoder(rec.Body).Decode(&tokens)
	s.Require().NoError(err)
	s.Require().Len(tokens, 0)
}

// TestListInvalidLimit tests list with invalid limit parameter
func (s *TokenHandlerTestSuite) TestListInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "101")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidContract tests handling of invalid contract address
func (s *TokenHandlerTestSuite) TestListInvalidContract() {
	q := make(url.Values)
	q.Set("contract", "invalid_address")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// ====================================
// Token Get Tests
// ====================================

// TestGetSuccess tests successful retrieval of a token
func (s *TokenHandlerTestSuite) TestGetSuccess() {
	contractAddr := "0x1234567890123456789012345678901234567890"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token/:contract/:token_id")
	c.SetParamNames("contract", "token_id")
	c.SetParamValues(contractAddr, "0")

	hashBytes, err := pkgTypes.HexFromString(contractAddr)
	s.Require().NoError(err)

	s.address.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	s.token.EXPECT().
		Get(gomock.Any(), uint64(1), decimal.NewFromInt(0)).
		Return(testToken1, nil).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var token responses.Token
	err = json.NewDecoder(rec.Body).Decode(&token)
	s.Require().NoError(err)
	s.Require().EqualValues(1, token.Id)
	s.Require().Equal("Test Token", token.Name)
	s.Require().Equal("TST", token.Symbol)
}

// TestGetNoContent tests when token is not found
func (s *TokenHandlerTestSuite) TestGetNoContent() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token/:contract/:token_id")
	c.SetParamNames("contract", "token_id")
	c.SetParamValues(testAddressHex1.Hex(), "999")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	s.token.EXPECT().
		Get(gomock.Any(), uint64(1), decimal.NewFromInt(999)).
		Return(storage.Token{}, sql.ErrNoRows).
		Times(1)

	s.token.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestGetInvalidContract tests handling of invalid contract address
func (s *TokenHandlerTestSuite) TestGetInvalidContract() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token/:contract/:token_id")
	c.SetParamNames("contract", "token_id")
	c.SetParamValues("invalid", "0")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestGetAddressNotFound tests when address is not found
func (s *TokenHandlerTestSuite) TestGetAddressNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token/:contract/:token_id")
	c.SetParamNames("contract", "token_id")
	c.SetParamValues(testAddressHex1.Hex(), "0")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{}, sql.ErrNoRows).
		Times(1)

	s.address.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// ====================================
// Transfer List Tests
// ====================================

// TestTransferListSuccess tests successful retrieval of transfers
func (s *TokenHandlerTestSuite) TestTransferListSuccess() {
	q := make(url.Values)
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer")

	s.transfer.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TransferListFilter) ([]storage.Transfer, error) {
			s.Require().Equal(10, filter.Limit)
			s.Require().Equal(0, filter.Offset)
			s.Require().Equal(sdk.SortOrderAsc, filter.Sort)
			return []storage.Transfer{testTransfer1, testTransfer2, testTransfer3}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TransferList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var transfers []responses.Transfer
	err := json.NewDecoder(rec.Body).Decode(&transfers)
	s.Require().NoError(err)
	s.Require().Len(transfers, 3)
}

// TestTransferListWithFilters tests transfer list with multiple filters
func (s *TokenHandlerTestSuite) TestTransferListWithFilters() {
	q := make(url.Values)
	q.Set("limit", "5")
	q.Set("height", "100")
	q.Set("type", "transfer,mint")
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer")

	s.transfer.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TransferListFilter) ([]storage.Transfer, error) {
			s.Require().Equal(5, filter.Limit)
			s.Require().Equal(0, filter.Offset)
			s.Require().NotNil(filter.Height)
			s.Require().EqualValues(100, *filter.Height)
			s.Require().Len(filter.Type, 2)
			return []storage.Transfer{testTransfer1, testTransfer2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TransferList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var transfers []responses.Transfer
	err := json.NewDecoder(rec.Body).Decode(&transfers)
	s.Require().NoError(err)
	s.Require().Len(transfers, 2)
}

// TestTransferListWithAddresses tests filtering by from/to addresses
func (s *TokenHandlerTestSuite) TestTransferListWithAddresses() {
	q := make(url.Values)
	q.Set("address_from", testAddressHex1.Hex())
	q.Set("address_to", testAddressHex2.Hex())
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex2).
		Return(storage.Address{Id: 2}, nil).
		Times(1)

	s.transfer.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TransferListFilter) ([]storage.Transfer, error) {
			s.Require().NotNil(filter.AddressFromId)
			s.Require().EqualValues(1, *filter.AddressFromId)
			s.Require().NotNil(filter.AddressToId)
			s.Require().EqualValues(2, *filter.AddressToId)
			return []storage.Transfer{testTransfer1}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TransferList(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestTransferListWithContract tests filtering by contract
func (s *TokenHandlerTestSuite) TestTransferListWithContract() {
	q := make(url.Values)
	q.Set("contract", testAddressHex1.Hex())
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	s.transfer.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TransferListFilter) ([]storage.Transfer, error) {
			s.Require().NotNil(filter.ContractId)
			s.Require().EqualValues(1, *filter.ContractId)
			return []storage.Transfer{testTransfer1, testTransfer2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TransferList(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestTransferListWithTimeRange tests filtering by time range
func (s *TokenHandlerTestSuite) TestTransferListWithTimeRange() {
	q := make(url.Values)
	q.Set("time_from", "1690855260")
	q.Set("time_to", "1690941660")
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer")

	s.transfer.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TransferListFilter) ([]storage.Transfer, error) {
			s.Require().False(filter.TimeFrom.IsZero())
			s.Require().False(filter.TimeTo.IsZero())
			return []storage.Transfer{testTransfer1, testTransfer2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TransferList(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestTransferListInvalidAddressFrom tests handling of invalid from address
func (s *TokenHandlerTestSuite) TestTransferListInvalidAddressFrom() {
	q := make(url.Values)
	q.Set("address_from", "invalid")
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer")

	s.Require().NoError(s.handler.TransferList(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// ====================================
// Transfer Get Tests
// ====================================

// TestGetTransferSuccess tests successful retrieval of a transfer
func (s *TokenHandlerTestSuite) TestGetTransferSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	s.transfer.EXPECT().
		Get(gomock.Any(), uint64(1)).
		Return(testTransfer1, nil).
		Times(1)

	s.Require().NoError(s.handler.GetTransfer(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var transfer responses.Transfer
	err := json.NewDecoder(rec.Body).Decode(&transfer)
	s.Require().NoError(err)
	s.Require().EqualValues(1, transfer.Id)
}

// TestGetTransferNoContent tests when transfer is not found
func (s *TokenHandlerTestSuite) TestGetTransferNoContent() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	s.transfer.EXPECT().
		Get(gomock.Any(), uint64(999)).
		Return(storage.Transfer{}, sql.ErrNoRows).
		Times(1)

	s.transfer.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.GetTransfer(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestGetTransferInvalidId tests handling of invalid transfer ID
func (s *TokenHandlerTestSuite) TestGetTransferInvalidId() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer/:id")
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	s.Require().NoError(s.handler.GetTransfer(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// TestGetTransferZeroId tests handling of zero transfer ID
func (s *TokenHandlerTestSuite) TestGetTransferZeroId() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/transfer/:id")
	c.SetParamNames("id")
	c.SetParamValues("0")

	s.Require().NoError(s.handler.GetTransfer(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// ====================================
// Token Balance List Tests
// ====================================

// TestTokenBalanceListSuccess tests successful retrieval of token balances
func (s *TokenHandlerTestSuite) TestTokenBalanceListSuccess() {
	q := make(url.Values)
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.tbs.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TokenBalanceListFilter) ([]storage.TokenBalance, error) {
			s.Require().Equal(10, filter.Limit)
			s.Require().Equal(0, filter.Offset)
			s.Require().Equal(sdk.SortOrderAsc, filter.Sort)
			return []storage.TokenBalance{testTokenBalance1, testTokenBalance2, testTokenBalance3}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var balances []responses.TokenBalance
	err := json.NewDecoder(rec.Body).Decode(&balances)
	s.Require().NoError(err)
	s.Require().Len(balances, 3)
}

// TestTokenBalanceListWithAddress tests filtering by address
func (s *TokenHandlerTestSuite) TestTokenBalanceListWithAddress() {
	q := make(url.Values)
	q.Set("address", testAddressHex1.Hex())
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	s.tbs.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TokenBalanceListFilter) ([]storage.TokenBalance, error) {
			s.Require().NotNil(filter.AddressId)
			s.Require().EqualValues(1, *filter.AddressId)
			return []storage.TokenBalance{testTokenBalance1, testTokenBalance3}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var balances []responses.TokenBalance
	err := json.NewDecoder(rec.Body).Decode(&balances)
	s.Require().NoError(err)
	s.Require().Len(balances, 2)
}

// TestTokenBalanceListWithContract tests filtering by contract
func (s *TokenHandlerTestSuite) TestTokenBalanceListWithContract() {
	q := make(url.Values)
	q.Set("contract", testAddressHex1.Hex())
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	s.tbs.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TokenBalanceListFilter) ([]storage.TokenBalance, error) {
			s.Require().NotNil(filter.ContractId)
			s.Require().EqualValues(1, *filter.ContractId)
			return []storage.TokenBalance{testTokenBalance1, testTokenBalance2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var balances []responses.TokenBalance
	err := json.NewDecoder(rec.Body).Decode(&balances)
	s.Require().NoError(err)
	s.Require().Len(balances, 2)
}

// TestTokenBalanceListWithTokenId tests filtering by token ID
func (s *TokenHandlerTestSuite) TestTokenBalanceListWithTokenId() {
	q := make(url.Values)
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.tbs.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.TokenBalanceListFilter) ([]storage.TokenBalance, error) {
			s.Require().NotNil(filter.TokenId)
			s.Require().Equal("0", filter.TokenId.String())
			return []storage.TokenBalance{testTokenBalance1}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var balances []responses.TokenBalance
	err := json.NewDecoder(rec.Body).Decode(&balances)
	s.Require().NoError(err)
	s.Require().Len(balances, 1)
}

// TestTokenBalanceListInvalidAddress tests handling of invalid address
func (s *TokenHandlerTestSuite) TestTokenBalanceListInvalidAddress() {
	q := make(url.Values)
	q.Set("address", "invalid")
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// TestTokenBalanceListInvalidContract tests handling of invalid contract
func (s *TokenHandlerTestSuite) TestTokenBalanceListInvalidContract() {
	q := make(url.Values)
	q.Set("contract", "invalid")
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// TestTokenBalanceListEmptyResult tests when no balances are found
func (s *TokenHandlerTestSuite) TestTokenBalanceListEmptyResult() {
	q := make(url.Values)
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.tbs.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		Return([]storage.TokenBalance{}, nil).
		Times(1)

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var balances []responses.TokenBalance
	err := json.NewDecoder(rec.Body).Decode(&balances)
	s.Require().NoError(err)
	s.Require().Len(balances, 0)
}

// TestTokenBalanceListInvalidLimit tests handling of invalid limit
func (s *TokenHandlerTestSuite) TestTokenBalanceListInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "101")
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// TestTokenBalanceListNegativeOffset tests handling of negative offset
func (s *TokenHandlerTestSuite) TestTokenBalanceListNegativeOffset() {
	q := make(url.Values)
	q.Set("offset", "-1")
	q.Set("token_id", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/token_balance")

	s.Require().NoError(s.handler.TokenBalanceList(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

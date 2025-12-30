package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/mock"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var (
	testAddressHex1 = pkgTypes.Hex{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13}
	testAddressHex2 = pkgTypes.Hex{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x14}
	testAddressHex3 = pkgTypes.Hex{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x15}

	testAddress1 = storage.Address{
		Id:             1,
		FirstHeight:    100,
		LastHeight:     200,
		Hash:           testAddressHex1,
		IsContract:     false,
		TxsCount:       50,
		ContractsCount: 0,
		Interactions:   75,
		Balance: &storage.Balance{
			Id:    1,
			Value: decimal.RequireFromString("1000000000"),
		},
	}

	testAddress2 = storage.Address{
		Id:             2,
		FirstHeight:    101,
		LastHeight:     201,
		Hash:           testAddressHex2,
		IsContract:     true,
		TxsCount:       100,
		ContractsCount: 5,
		Interactions:   150,
		Balance: &storage.Balance{
			Id:    2,
			Value: decimal.RequireFromString("5000000000"),
		},
	}

	testAddress3 = storage.Address{
		Id:             3,
		FirstHeight:    102,
		LastHeight:     202,
		Hash:           testAddressHex3,
		IsContract:     true,
		TxsCount:       25,
		ContractsCount: 2,
		Interactions:   30,
		Balance: &storage.Balance{
			Id:    3,
			Value: decimal.RequireFromString("2500000000"),
		},
	}
)

// AddressHandlerTestSuite -
type AddressHandlerTestSuite struct {
	suite.Suite
	address *mock.MockIAddress
	echo    *echo.Echo
	handler *AddressHandler
	ctrl    *gomock.Controller
}

// SetupSuite -
func (s *AddressHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.address = mock.NewMockIAddress(s.ctrl)
	s.handler = NewAddressHandler(s.address)
}

// TearDownSuite -
func (s *AddressHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteAddressHandler_Run(t *testing.T) {
	suite.Run(t, new(AddressHandlerTestSuite))
}

// ====================================
// Address List Tests
// ====================================

// TestListSuccess tests successful retrieval of addresses with default parameters
func (s *AddressHandlerTestSuite) TestListSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address")

	s.address.EXPECT().
		ListWithBalance(gomock.Any(), storage.AddressListFilter{
			Limit:         10,
			Offset:        0,
			Sort:          sdk.SortOrderAsc,
			SortField:     "",
			OnlyContracts: false,
		}).
		Return([]storage.Address{testAddress1, testAddress2, testAddress3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var addresses []responses.Address
	err := json.NewDecoder(rec.Body).Decode(&addresses)
	s.Require().NoError(err)
	s.Require().Len(addresses, 3)

	s.Require().EqualValues(1, addresses[0].Id)
	s.Require().Equal(testAddress1.Hash.Hex(), addresses[0].Hash)
	s.Require().False(addresses[0].IsContract)
	s.Require().EqualValues(50, addresses[0].TxCount)
}

// TestListWithLimit tests list with custom limit parameter
func (s *AddressHandlerTestSuite) TestListWithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address")

	s.address.EXPECT().
		ListWithBalance(gomock.Any(), storage.AddressListFilter{
			Limit:         5,
			Offset:        0,
			Sort:          sdk.SortOrderAsc,
			SortField:     "",
			OnlyContracts: false,
		}).
		Return([]storage.Address{testAddress1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var addresses []responses.Address
	err := json.NewDecoder(rec.Body).Decode(&addresses)
	s.Require().NoError(err)
	s.Require().Len(addresses, 1)
}

// TestListWithOnlyContracts tests filtering only contract addresses
func (s *AddressHandlerTestSuite) TestListWithOnlyContracts() {
	q := make(url.Values)
	q.Set("only_contracts", "true")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address")

	s.address.EXPECT().
		ListWithBalance(gomock.Any(), storage.AddressListFilter{
			Limit:         10,
			Offset:        0,
			Sort:          sdk.SortOrderAsc,
			SortField:     "",
			OnlyContracts: true,
		}).
		Return([]storage.Address{testAddress2, testAddress3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var addresses []responses.Address
	err := json.NewDecoder(rec.Body).Decode(&addresses)
	s.Require().NoError(err)
	s.Require().Len(addresses, 2)
	s.Require().True(addresses[0].IsContract)
	s.Require().True(addresses[1].IsContract)
}

// TestListWithSortBy tests list with sort_by parameter
func (s *AddressHandlerTestSuite) TestListWithSortBy() {
	q := make(url.Values)
	q.Set("sort_by", "last_height")
	q.Set("sort", "desc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address")

	s.address.EXPECT().
		ListWithBalance(gomock.Any(), storage.AddressListFilter{
			Limit:         10,
			Offset:        0,
			Sort:          sdk.SortOrderDesc,
			SortField:     "last_height",
			OnlyContracts: false,
		}).
		Return([]storage.Address{testAddress3, testAddress2, testAddress1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var addresses []responses.Address
	err := json.NewDecoder(rec.Body).Decode(&addresses)
	s.Require().NoError(err)
	s.Require().Len(addresses, 3)
}

// TestListInvalidLimit tests list with invalid limit parameter
func (s *AddressHandlerTestSuite) TestListInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "101")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListEmptyResult tests list when no addresses are found
func (s *AddressHandlerTestSuite) TestListEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address")

	s.address.EXPECT().
		ListWithBalance(gomock.Any(), gomock.Any()).
		Return([]storage.Address{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var addresses []responses.Address
	err := json.NewDecoder(rec.Body).Decode(&addresses)
	s.Require().NoError(err)
	s.Require().Len(addresses, 0)
}

// ====================================
// Address Get Tests
// ====================================

// TestGetSuccess tests successful retrieval of an address
func (s *AddressHandlerTestSuite) TestGetSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testAddress1, nil).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var address responses.Address
	err := json.NewDecoder(rec.Body).Decode(&address)
	s.Require().NoError(err)
	s.Require().EqualValues(1, address.Id)
	s.Require().Equal(testAddress1.Hash.Hex(), address.Hash)
	s.Require().False(address.IsContract)
	s.Require().EqualValues(50, address.TxCount)
}

// TestGetNoContent tests when address is not found
func (s *AddressHandlerTestSuite) TestGetNoContent() {
	addressHash := "0xaabbccddee123456789012345678901234567890"
	hashBytes, err := pkgTypes.HexFromString(addressHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(addressHash)

	s.address.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Address{}, sql.ErrNoRows).
		Times(1)

	s.address.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestGetInvalidHash tests handling of invalid address hash
func (s *AddressHandlerTestSuite) TestGetInvalidHash() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address/:hash")
	c.SetParamNames("hash")
	c.SetParamValues("invalid_hash")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestGetInvalidHashLength tests handling of invalid hash length
func (s *AddressHandlerTestSuite) TestGetInvalidHashLength() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/address/:hash")
	c.SetParamNames("hash")
	c.SetParamValues("0x01")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

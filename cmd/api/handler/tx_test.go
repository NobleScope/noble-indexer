package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/mock"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var (
	testTxHash = pkgTypes.Hex{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11,
		0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
	}
)

// TxHandlerTestSuite -
type TxHandlerTestSuite struct {
	suite.Suite
	tx      *mock.MockITx
	trace   *mock.MockITrace
	address *mock.MockIAddress
	echo    *echo.Echo
	handler *TxHandler
	ctrl    *gomock.Controller
}

// SetupSuite -
func (s *TxHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.tx = mock.NewMockITx(s.ctrl)
	s.trace = mock.NewMockITrace(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.handler = NewTxHandler(s.tx, s.trace, s.address, testIndexerName)
}

// TearDownSuite -
func (s *TxHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteTxHandler_Run(t *testing.T) {
	suite.Run(t, new(TxHandlerTestSuite))
}

// TestGetSuccess tests successful retrieval of a transaction with to address
func (s *TxHandlerTestSuite) TestGetSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(testTxHash.Hex())

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(testTxWithToAddress, nil).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tx responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&tx)
	s.Require().NoError(err)
	s.Require().EqualValues(100, tx.Height)
	s.Require().Equal("0x010203", tx.Hash)
	s.Require().EqualValues(0, tx.Index)
	s.Require().Equal(testAddressHex1.Hex(), tx.FromAddress)
	s.Require().NotNil(tx.ToAddress)
	s.Require().Equal(testAddressHex2.Hex(), *tx.ToAddress)
	s.Require().Equal("1000000000000000000", tx.Amount.String())
	s.Require().Equal("21000", tx.Gas.String())
	s.Require().Equal("1000000", tx.GasPrice.String())
	s.Require().Equal("21000000000", tx.Fee.String())
	s.Require().Equal("TxStatusSuccess", tx.Status)
}

// TestGetContractCreation tests retrieval of a contract creation transaction
func (s *TxHandlerTestSuite) TestGetContractCreation() {
	txHash := "0x0405060000000000000000000000000000000000000000000000000000000000"
	hashBytes, err := pkgTypes.HexFromString(txHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(txHash)

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTxContractCreation, nil).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tx responses.Transaction
	err = json.NewDecoder(rec.Body).Decode(&tx)
	s.Require().NoError(err)
	s.Require().EqualValues(100, tx.Height)
	s.Require().Equal("0x040506", tx.Hash)
	s.Require().EqualValues(1, tx.Index)
	s.Require().Equal(testAddressHex1.Hex(), tx.FromAddress)
	s.Require().Nil(tx.ToAddress)
	s.Require().Equal("0", tx.Amount.String())
	s.Require().Equal("100000", tx.Gas.String())
	s.Require().Equal("2000000", tx.GasPrice.String())
	s.Require().Equal("200000000000", tx.Fee.String())
	s.Require().Equal("TxStatusSuccess", tx.Status)
}

// TestGetContractCall tests retrieval of a contract call transaction
func (s *TxHandlerTestSuite) TestGetContractCall() {
	txHash := "0x0708090000000000000000000000000000000000000000000000000000000000"
	hashBytes, err := pkgTypes.HexFromString(txHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(txHash)

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTxContractCall, nil).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tx responses.Transaction
	err = json.NewDecoder(rec.Body).Decode(&tx)
	s.Require().NoError(err)
	s.Require().EqualValues(100, tx.Height)
	s.Require().Equal("0x070809", tx.Hash)
	s.Require().EqualValues(2, tx.Index)
	s.Require().Equal(testAddressHex1.Hex(), tx.FromAddress)
	s.Require().NotNil(tx.ToAddress)
	s.Require().Equal(testAddressHex2.Hex(), *tx.ToAddress)
	s.Require().Equal("100000000", tx.Amount.String())
	s.Require().Equal("50000", tx.Gas.String())
	s.Require().Equal("1500000", tx.GasPrice.String())
	s.Require().Equal("75000000000", tx.Fee.String())
	s.Require().Equal("TxStatusSuccess", tx.Status)
}

// TestGetNoContent tests when transaction is not found
func (s *TxHandlerTestSuite) TestGetNoContent() {
	txHash := "0xaabbccddee000000000000000000000000000000000000000000000000000000"
	hashBytes, err := pkgTypes.HexFromString(txHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(txHash)

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Tx{}, sql.ErrNoRows).
		Times(1)

	s.tx.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestGetInvalidHash tests handling of invalid transaction hash
func (s *TxHandlerTestSuite) TestGetInvalidHash() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues("invalid_hash")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestGetMissingHash tests handling of missing hash parameter
func (s *TxHandlerTestSuite) TestGetMissingHash() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues("")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestGetInvalidHashLength tests handling of invalid hash length
func (s *TxHandlerTestSuite) TestGetInvalidHashLength() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues("0x01")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestGetHashWithoutPrefix tests handling of hash without 0x prefix
func (s *TxHandlerTestSuite) TestGetHashWithoutPrefix() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues("010203")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListSuccess tests successful retrieval of transactions list
func (s *TxHandlerTestSuite) TestListSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{
			testTxWithToAddress,
			testTxContractCreation,
			testTxContractCall,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 3)

	s.Require().EqualValues(100, txs[0].Height)
	s.Require().Equal("0x010203", txs[0].Hash)
	s.Require().EqualValues(0, txs[0].Index)
	s.Require().Equal(testAddressHex1.Hex(), txs[0].FromAddress)
	s.Require().NotNil(txs[0].ToAddress)
	s.Require().Equal(testAddressHex2.Hex(), *txs[0].ToAddress)

	s.Require().EqualValues(100, txs[1].Height)
	s.Require().Equal("0x040506", txs[1].Hash)
	s.Require().EqualValues(1, txs[1].Index)
	s.Require().Nil(txs[1].ToAddress)

	s.Require().EqualValues(100, txs[2].Height)
	s.Require().Equal("0x070809", txs[2].Hash)
	s.Require().EqualValues(2, txs[2].Index)
	s.Require().NotNil(txs[2].ToAddress)
}

// TestListWithLimit tests list with custom limit parameter
func (s *TxHandlerTestSuite) TestListWithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  5,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{
			testTxWithToAddress,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 1)
}

// TestListWithOffset tests list with offset parameter
func (s *TxHandlerTestSuite) TestListWithOffset() {
	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  10,
			Offset: 5,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{
			testTxContractCall,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 1)
	s.Require().EqualValues(2, txs[0].Index)
}

// TestListWithLimitAndOffset tests list with both limit and offset
func (s *TxHandlerTestSuite) TestListWithLimitAndOffset() {
	q := make(url.Values)
	q.Set("limit", "2")
	q.Set("offset", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  2,
			Offset: 1,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{
			testTxContractCreation,
			testTxContractCall,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 2)
	s.Require().EqualValues(1, txs[0].Index)
	s.Require().EqualValues(2, txs[1].Index)
}

// TestListAscOrder tests list with ascending sort order
func (s *TxHandlerTestSuite) TestListAscOrder() {
	q := make(url.Values)
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{
			testTxWithToAddress,
			testTxContractCreation,
			testTxContractCall,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 3)
	s.Require().EqualValues(0, txs[0].Index)
	s.Require().EqualValues(1, txs[1].Index)
	s.Require().EqualValues(2, txs[2].Index)
}

// TestListDescOrder tests list with descending sort order (default)
func (s *TxHandlerTestSuite) TestListDescOrder() {
	q := make(url.Values)
	q.Set("sort", "desc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{
			testTxContractCall,
			testTxContractCreation,
			testTxWithToAddress,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 3)
	s.Require().EqualValues(2, txs[0].Index)
	s.Require().EqualValues(1, txs[1].Index)
	s.Require().EqualValues(0, txs[2].Index)
}

// TestListEmptyResult tests list when no transactions are found
func (s *TxHandlerTestSuite) TestListEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 0)
}

// TestListInvalidLimit tests list with invalid limit parameter
func (s *TxHandlerTestSuite) TestListInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TxType{},
			Status: []types.TxStatus{},
		}).
		Return([]storage.Tx{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestListMaxLimit tests list with limit exceeding maximum
func (s *TxHandlerTestSuite) TestListMaxLimit() {
	q := make(url.Values)
	q.Set("limit", "101")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidSort tests list with invalid sort parameter
func (s *TxHandlerTestSuite) TestListInvalidSort() {
	q := make(url.Values)
	q.Set("sort", "invalid")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListNegativeOffset tests list with negative offset
func (s *TxHandlerTestSuite) TestListNegativeOffset() {
	q := make(url.Values)
	q.Set("offset", "-1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

func (s *TxHandlerTestSuite) TestListByAddressId() {
	q := make(url.Values)
	q.Set("address", testAddressHex1.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testFromAddress, nil).
		Times(1)

	addressId := testFromAddress.Id
	s.tx.EXPECT().
		Filter(gomock.Any(), storage.TxListFilter{
			Limit:     10,
			Offset:    0,
			Sort:      sdk.SortOrderDesc,
			Type:      []types.TxType{},
			Status:    []types.TxStatus{},
			AddressId: &addressId,
		}).
		Return([]storage.Tx{
			testTxWithToAddress,
			testTxContractCreation,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 2)
}

// ====================================
// Traces Tests
// ====================================

var (
	testTrace1 = storage.Trace{
		Id:           1,
		Height:       100,
		Time:         testTxWithToAddress.Time,
		TxId:         uint64Ptr(1),
		From:         uint64Ptr(1),
		To:           uint64Ptr(2),
		GasLimit:     decimal.NewFromInt(21000),
		Amount:       &testTxWithToAddress.Amount,
		Input:        []byte{},
		TxPosition:   uint64Ptr(0),
		TraceAddress: []uint64{},
		Type:         types.Call,
		GasUsed:      decimal.NewFromInt(21000),
		Output:       []byte{},
		Subtraces:    0,
		FromAddress:  &testFromAddress,
		ToAddress:    &testToAddress,
		Tx:           &testTxWithToAddress,
	}

	testTrace2 = storage.Trace{
		Id:             2,
		Height:         100,
		Time:           testTxContractCreation.Time,
		TxId:           uint64Ptr(2),
		From:           uint64Ptr(1),
		To:             nil,
		GasLimit:       decimal.NewFromInt(100000),
		Amount:         nil,
		Input:          []byte{0x60, 0x60, 0x60},
		TxPosition:     uint64Ptr(1),
		TraceAddress:   []uint64{},
		Type:           types.Create,
		GasUsed:        decimal.NewFromInt(100000),
		Output:         []byte{0x60, 0x60, 0x60},
		Subtraces:      0,
		FromAddress:    &testFromAddress,
		ToAddress:      nil,
		Tx:             &testTxContractCreation,
		CreationMethod: stringPtr("create"),
	}

	testTrace3 = storage.Trace{
		Id:           3,
		Height:       100,
		Time:         testTxContractCall.Time,
		TxId:         uint64Ptr(3),
		From:         uint64Ptr(1),
		To:           uint64Ptr(2),
		GasLimit:     decimal.NewFromInt(50000),
		Amount:       &testTxContractCall.Amount,
		Input:        []byte{0xa9, 0x05, 0x9c, 0xbb},
		TxPosition:   uint64Ptr(2),
		TraceAddress: []uint64{},
		Type:         types.Call,
		GasUsed:      decimal.NewFromInt(50000),
		Output:       []byte{0x00, 0x01},
		Subtraces:    0,
		FromAddress:  &testFromAddress,
		ToAddress:    &testToAddress,
		Tx:           &testTxContractCall,
	}
)

func stringPtr(s string) *string {
	return &s
}

// TestTracesSuccess tests successful retrieval of traces with default parameters
func (s *TxHandlerTestSuite) TestTracesSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TraceType{},
		}).
		Return([]*storage.Trace{&testTrace1, &testTrace2, &testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 3)

	s.Require().EqualValues(100, traces[0].Height)
	s.Require().NotNil(traces[0].TxHash)
	s.Require().Equal("0x010203", *traces[0].TxHash)
	s.Require().Equal("call", traces[0].Type)
}

// TestTracesWithTxHash tests traces filtered by transaction hash
func (s *TxHandlerTestSuite) TestTracesWithTxHash() {
	q := make(url.Values)
	q.Set("tx_hash", testTxHash.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(testTxWithToAddress, nil).
		Times(1)

	txId := testTxWithToAddress.Id
	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TraceType{},
			TxId:   &txId,
		}).
		Return([]*storage.Trace{&testTrace1}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 1)
	s.Require().NotNil(traces[0].TxHash)
	s.Require().Equal("0x010203", *traces[0].TxHash)
}

// TestTracesWithAddressFrom tests traces filtered by from address
func (s *TxHandlerTestSuite) TestTracesWithAddressFrom() {
	q := make(url.Values)
	q.Set("address_from", testFromAddress.Hash.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.address.EXPECT().
		ByHash(gomock.Any(), testFromAddress.Hash).
		Return(testFromAddress, nil).
		Times(1)

	fromId := testFromAddress.Id
	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:         10,
			Offset:        0,
			Sort:          sdk.SortOrderDesc,
			Type:          []types.TraceType{},
			AddressFromId: &fromId,
		}).
		Return([]*storage.Trace{&testTrace1, &testTrace2, &testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 3)
}

// TestTracesWithAddressTo tests traces filtered by to address
func (s *TxHandlerTestSuite) TestTracesWithAddressTo() {
	q := make(url.Values)
	q.Set("address_to", testToAddress.Hash.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.address.EXPECT().
		ByHash(gomock.Any(), testToAddress.Hash).
		Return(testToAddress, nil).
		Times(1)

	toId := testToAddress.Id
	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:       10,
			Offset:      0,
			Sort:        sdk.SortOrderDesc,
			Type:        []types.TraceType{},
			AddressToId: &toId,
		}).
		Return([]*storage.Trace{&testTrace1, &testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 2)
}

func (s *TxHandlerTestSuite) TestTracesWithAddress() {
	q := make(url.Values)
	q.Set("address", testToAddress.Hash.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.address.EXPECT().
		ByHash(gomock.Any(), testToAddress.Hash).
		Return(testToAddress, nil).
		Times(1)

	addressId := testToAddress.Id
	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:     10,
			Offset:    0,
			Sort:      sdk.SortOrderDesc,
			Type:      []types.TraceType{},
			AddressId: &addressId,
		}).
		Return([]*storage.Trace{&testTrace1, &testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 2)
}

// TestTracesWithType tests traces filtered by type
func (s *TxHandlerTestSuite) TestTracesWithType() {
	q := make(url.Values)
	q.Set("type", "call,create")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TraceType{types.Call, types.Create},
		}).
		Return([]*storage.Trace{&testTrace1, &testTrace2, &testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 3)
}

// TestTracesWithHeight tests traces filtered by block height
func (s *TxHandlerTestSuite) TestTracesWithHeight() {
	q := make(url.Values)
	q.Set("height", "100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	height := uint64(100)
	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TraceType{},
			Height: &height,
		}).
		Return([]*storage.Trace{&testTrace1, &testTrace2, &testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 3)
}

// TestTracesWithLimitAndOffset tests traces with custom limit and offset
func (s *TxHandlerTestSuite) TestTracesWithLimitAndOffset() {
	q := make(url.Values)
	q.Set("limit", "5")
	q.Set("offset", "2")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:  5,
			Offset: 2,
			Sort:   sdk.SortOrderDesc,
			Type:   []types.TraceType{},
		}).
		Return([]*storage.Trace{&testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 1)
}

// TestTracesWithSortAsc tests traces with ascending sort order
func (s *TxHandlerTestSuite) TestTracesWithSortAsc() {
	q := make(url.Values)
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.trace.EXPECT().
		Filter(gomock.Any(), storage.TraceListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
			Type:   []types.TraceType{},
		}).
		Return([]*storage.Trace{&testTrace1, &testTrace2, &testTrace3}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 3)
}

// TestTracesEmptyResult tests when no traces are found
func (s *TxHandlerTestSuite) TestTracesEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.trace.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		Return([]*storage.Trace{}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err := json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 0)
}

// ====================================
// TxTracesTree Tests
// ====================================

var (
	testTraceRoot = storage.Trace{
		Id:           10,
		Height:       100,
		Time:         testTxWithToAddress.Time,
		TxId:         uint64Ptr(1),
		From:         uint64Ptr(1),
		To:           uint64Ptr(2),
		GasLimit:     decimal.NewFromInt(100000),
		Amount:       &testTxWithToAddress.Amount,
		Input:        []byte{0xa9, 0x05, 0x9c, 0xbb},
		TxPosition:   uint64Ptr(0),
		TraceAddress: []uint64{},
		Type:         types.Call,
		GasUsed:      decimal.NewFromInt(80000),
		Output:       []byte{0x00, 0x01},
		Subtraces:    2,
		FromAddress:  &testFromAddress,
		ToAddress:    &testToAddress,
		Tx:           &testTxWithToAddress,
	}

	testTraceChild0 = storage.Trace{
		Id:           11,
		Height:       100,
		Time:         testTxWithToAddress.Time,
		TxId:         uint64Ptr(1),
		From:         uint64Ptr(2),
		To:           uint64Ptr(1),
		GasLimit:     decimal.NewFromInt(50000),
		Amount:       nil,
		Input:        []byte{0x70, 0xa0, 0x82, 0x31},
		TxPosition:   uint64Ptr(0),
		TraceAddress: []uint64{0},
		Type:         types.Call,
		GasUsed:      decimal.NewFromInt(30000),
		Output:       []byte{0x00},
		Subtraces:    1,
		FromAddress:  &testToAddress,
		ToAddress:    &testFromAddress,
		Tx:           &testTxWithToAddress,
	}

	testTraceChild1 = storage.Trace{
		Id:           12,
		Height:       100,
		Time:         testTxWithToAddress.Time,
		TxId:         uint64Ptr(1),
		From:         uint64Ptr(2),
		To:           uint64Ptr(1),
		GasLimit:     decimal.NewFromInt(30000),
		Amount:       nil,
		Input:        []byte{},
		TxPosition:   uint64Ptr(0),
		TraceAddress: []uint64{1},
		Type:         types.Delegatecall,
		GasUsed:      decimal.NewFromInt(20000),
		Output:       []byte{},
		Subtraces:    0,
		FromAddress:  &testToAddress,
		ToAddress:    &testFromAddress,
		Tx:           &testTxWithToAddress,
	}

	testTraceGrandchild = storage.Trace{
		Id:           13,
		Height:       100,
		Time:         testTxWithToAddress.Time,
		TxId:         uint64Ptr(1),
		From:         uint64Ptr(1),
		To:           uint64Ptr(2),
		GasLimit:     decimal.NewFromInt(10000),
		Amount:       nil,
		Input:        []byte{0x01, 0x02},
		TxPosition:   uint64Ptr(0),
		TraceAddress: []uint64{0, 0},
		Type:         types.Staticcall,
		GasUsed:      decimal.NewFromInt(5000),
		Output:       []byte{0x01},
		Subtraces:    0,
		FromAddress:  &testFromAddress,
		ToAddress:    &testToAddress,
		Tx:           &testTxWithToAddress,
	}
)

// TestTxTracesTreeSuccess tests successful retrieval of a flat trace tree (single root)
func (s *TxHandlerTestSuite) TestTxTracesTreeSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/txs/:hash/traces_tree")
	c.SetParamNames("hash")
	c.SetParamValues(testTxHash.Hex())

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(testTxWithToAddress, nil).
		Times(1)

	s.trace.EXPECT().
		ByTxId(gomock.Any(), testTxWithToAddress.Id).
		Return([]*storage.Trace{&testTrace1}, nil).
		Times(1)

	s.Require().NoError(s.handler.TxTracesTree(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tree responses.TraceTreeItem
	err := json.NewDecoder(rec.Body).Decode(&tree)
	s.Require().NoError(err)
	s.Require().NotNil(tree.Trace)
	s.Require().Equal("call", tree.Type)
	s.Require().Len(tree.Children, 0)
}

// TestTxTracesTreeWithHierarchy tests trace tree with nested children
func (s *TxHandlerTestSuite) TestTxTracesTreeWithHierarchy() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/txs/:hash/traces_tree")
	c.SetParamNames("hash")
	c.SetParamValues(testTxHash.Hex())

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(testTxWithToAddress, nil).
		Times(1)

	s.trace.EXPECT().
		ByTxId(gomock.Any(), testTxWithToAddress.Id).
		Return([]*storage.Trace{
			&testTraceRoot,
			&testTraceChild0,
			&testTraceGrandchild,
			&testTraceChild1,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.TxTracesTree(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tree responses.TraceTreeItem
	err := json.NewDecoder(rec.Body).Decode(&tree)
	s.Require().NoError(err)

	// root
	s.Require().NotNil(tree.Trace)
	s.Require().Equal("call", tree.Type)

	// root has 2 children
	s.Require().Len(tree.Children, 2)
	s.Require().Equal("call", tree.Children[0].Type)
	s.Require().Equal("delegatecall", tree.Children[1].Type)

	// first child has 1 grandchild
	s.Require().Len(tree.Children[0].Children, 1)
	s.Require().Equal("staticcall", tree.Children[0].Children[0].Type)
	s.Require().Len(tree.Children[0].Children[0].Children, 0)

	// second child has no children
	s.Require().Len(tree.Children[1].Children, 0)
}

// TestTxTracesTreeEmptyTraces tests when transaction has no traces
func (s *TxHandlerTestSuite) TestTxTracesTreeEmptyTraces() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/txs/:hash/traces_tree")
	c.SetParamNames("hash")
	c.SetParamValues(testTxHash.Hex())

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(testTxWithToAddress, nil).
		Times(1)

	s.trace.EXPECT().
		ByTxId(gomock.Any(), testTxWithToAddress.Id).
		Return([]*storage.Trace{}, nil).
		Times(1)

	s.Require().NoError(s.handler.TxTracesTree(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tree *responses.TraceTreeItem
	err := json.NewDecoder(rec.Body).Decode(&tree)
	s.Require().NoError(err)
	s.Require().Nil(tree)
}

// TestTxTracesTreeTxNotFound tests when transaction is not found
func (s *TxHandlerTestSuite) TestTxTracesTreeTxNotFound() {
	txHash := "0xaabbccddee000000000000000000000000000000000000000000000000000000"
	hashBytes, err := pkgTypes.HexFromString(txHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/txs/:hash/traces_tree")
	c.SetParamNames("hash")
	c.SetParamValues(txHash)

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Tx{}, sql.ErrNoRows).
		Times(1)

	s.tx.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.TxTracesTree(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestTxTracesTreeInvalidHash tests handling of invalid transaction hash
func (s *TxHandlerTestSuite) TestTxTracesTreeInvalidHash() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/txs/:hash/traces_tree")
	c.SetParamNames("hash")
	c.SetParamValues("invalid_hash")

	s.Require().NoError(s.handler.TxTracesTree(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestTxTracesTreeMissingHash tests handling of missing hash parameter
func (s *TxHandlerTestSuite) TestTxTracesTreeMissingHash() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/txs/:hash/traces_tree")
	c.SetParamNames("hash")
	c.SetParamValues("")

	s.Require().NoError(s.handler.TxTracesTree(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestTxTracesTreeByTxIdError tests error from ByTxId storage call
func (s *TxHandlerTestSuite) TestTxTracesTreeByTxIdError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/txs/:hash/traces_tree")
	c.SetParamNames("hash")
	c.SetParamValues(testTxHash.Hex())

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(testTxWithToAddress, nil).
		Times(1)

	s.trace.EXPECT().
		ByTxId(gomock.Any(), testTxWithToAddress.Id).
		Return(nil, sql.ErrNoRows).
		Times(1)

	s.trace.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.TxTracesTree(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestTracesInvalidLimit tests traces with invalid limit parameter
func (s *TxHandlerTestSuite) TestTracesInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "101")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

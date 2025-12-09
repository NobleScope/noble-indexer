// SPDX-FileCopyrightText: 2025 Bb Strategy Pte. Ltd. <celenium@baking-bad.org>
// SPDX-License-Identifier: MIT

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
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

const testTxHash = "0x0102030000000000000000000000000000000000000000000000000000000000"

// TxHandlerTestSuite -
type TxHandlerTestSuite struct {
	suite.Suite
	tx      *mock.MockITx
	trace   *mock.MockITrace
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
	s.handler = NewTxHandler(s.tx, s.trace, testIndexerName)
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
	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/tx/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(testTxHash)

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTxWithToAddress, nil).
		Times(1)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var tx responses.Transaction
	err = json.NewDecoder(rec.Body).Decode(&tx)
	s.Require().NoError(err)
	s.Require().EqualValues(100, tx.Height)
	s.Require().Equal("0x010203", tx.Hash)
	s.Require().EqualValues(0, tx.Index)
	s.Require().Equal("0x1234567890123456789012345678901234567890", tx.FromAddress)
	s.Require().NotNil(tx.ToAddress)
	s.Require().Equal("0x0987654321098765432109876543210987654321", *tx.ToAddress)
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
	s.Require().Equal("0x1234567890123456789012345678901234567890", tx.FromAddress)
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
	s.Require().Equal("0x1234567890123456789012345678901234567890", tx.FromAddress)
	s.Require().NotNil(tx.ToAddress)
	s.Require().Equal("0x0987654321098765432109876543210987654321", *tx.ToAddress)
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
		List(gomock.Any(), uint64(10), uint64(0), sdk.SortOrderDesc).
		Return([]*storage.Tx{
			&testTxWithToAddress,
			&testTxContractCreation,
			&testTxContractCall,
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
	s.Require().Equal("0x1234567890123456789012345678901234567890", txs[0].FromAddress)
	s.Require().NotNil(txs[0].ToAddress)
	s.Require().Equal("0x0987654321098765432109876543210987654321", *txs[0].ToAddress)

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
		List(gomock.Any(), uint64(5), uint64(0), sdk.SortOrderDesc).
		Return([]*storage.Tx{
			&testTxWithToAddress,
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
		List(gomock.Any(), uint64(10), uint64(5), sdk.SortOrderDesc).
		Return([]*storage.Tx{
			&testTxContractCall,
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
		List(gomock.Any(), uint64(2), uint64(1), sdk.SortOrderDesc).
		Return([]*storage.Tx{
			&testTxContractCreation,
			&testTxContractCall,
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
		List(gomock.Any(), uint64(10), uint64(0), sdk.SortOrderAsc).
		Return([]*storage.Tx{
			&testTxWithToAddress,
			&testTxContractCreation,
			&testTxContractCall,
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
		List(gomock.Any(), uint64(10), uint64(0), sdk.SortOrderDesc).
		Return([]*storage.Tx{
			&testTxContractCall,
			&testTxContractCreation,
			&testTxWithToAddress,
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
		List(gomock.Any(), uint64(10), uint64(0), sdk.SortOrderDesc).
		Return([]*storage.Tx{}, nil).
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
		List(gomock.Any(), uint64(10), uint64(0), sdk.SortOrderDesc).
		Return([]*storage.Tx{}, nil).
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

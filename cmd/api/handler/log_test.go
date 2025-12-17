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
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

const testAddressHash = "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"

var (
	testLog1 = storage.Log{
		Id:        1,
		Height:    100,
		Time:      testTime,
		Index:     0,
		Name:      "Transfer",
		TxId:      1,
		Data:      pkgTypes.Hex{0xFF, 0xFF, 0xFF},
		Topics:    []pkgTypes.Hex{{0xaa, 0xbb, 0xcc}},
		AddressId: 1,
		Removed:   false,
		Address: storage.Address{
			Id:      1,
			Address: testAddressHash,
		},
		Tx: storage.Tx{
			Id:   1,
			Hash: pkgTypes.Hex{0x01, 0x02, 0x03},
		},
	}

	testLog2 = storage.Log{
		Id:        2,
		Height:    101,
		Time:      testTime.Add(time.Hour),
		Index:     1,
		Name:      "Approval",
		TxId:      2,
		Data:      pkgTypes.Hex{0xAA, 0xBB, 0xCC},
		Topics:    []pkgTypes.Hex{{0x11, 0x22, 0x33}},
		AddressId: 2,
		Removed:   false,
		Address: storage.Address{
			Id:      2,
			Address: "0x1234567890123456789012345678901234567890",
		},
		Tx: storage.Tx{
			Id:   2,
			Hash: pkgTypes.Hex{0x04, 0x05, 0x06},
		},
	}

	testLog3 = storage.Log{
		Id:        3,
		Height:    102,
		Time:      testTime.Add(2 * time.Hour),
		Index:     0,
		Name:      "Swap",
		TxId:      3,
		Data:      pkgTypes.Hex{0x11, 0x22, 0x33},
		Topics:    []pkgTypes.Hex{{0xdd, 0xee, 0xff}},
		AddressId: 1,
		Removed:   false,
		Address: storage.Address{
			Id:      1,
			Address: testAddressHash,
		},
		Tx: storage.Tx{
			Id:   3,
			Hash: pkgTypes.Hex{0x07, 0x08, 0x09},
		},
	}
)

// LogHandlerTestSuite -
type LogHandlerTestSuite struct {
	suite.Suite
	log     *mock.MockILog
	tx      *mock.MockITx
	address *mock.MockIAddress
	echo    *echo.Echo
	handler *LogHandler
	ctrl    *gomock.Controller
}

// SetupSuite -
func (s *LogHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.log = mock.NewMockILog(s.ctrl)
	s.tx = mock.NewMockITx(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.handler = NewLogHandler(s.log, s.tx, s.address)
}

// TearDownSuite -
func (s *LogHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteLogHandler_Run(t *testing.T) {
	suite.Run(t, new(LogHandlerTestSuite))
}

// TestListSuccess tests successful retrieval of logs with default parameters
func (s *LogHandlerTestSuite) TestListSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.Log{testLog1, testLog2, testLog3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 3)

	s.Require().EqualValues(1, logs[0].Id)
	s.Require().EqualValues(100, logs[0].Height)
	s.Require().Equal(testTime, logs[0].Time)
	s.Require().EqualValues(0, logs[0].Index)
	s.Require().Equal("Transfer", logs[0].Name)
}

// TestListWithLimit tests list with custom limit parameter
func (s *LogHandlerTestSuite) TestListWithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  5,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.Log{testLog1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 1)
}

// TestListWithOffset tests list with offset parameter
func (s *LogHandlerTestSuite) TestListWithOffset() {
	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 5,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.Log{testLog2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 1)
	s.Require().EqualValues(2, logs[0].Id)
}

// TestListWithLimitAndOffset tests list with both limit and offset
func (s *LogHandlerTestSuite) TestListWithLimitAndOffset() {
	q := make(url.Values)
	q.Set("limit", "2")
	q.Set("offset", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  2,
			Offset: 1,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.Log{testLog2, testLog3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 2)
	s.Require().EqualValues(2, logs[0].Id)
	s.Require().EqualValues(3, logs[1].Id)
}

// TestListAscOrder tests list with ascending sort order
func (s *LogHandlerTestSuite) TestListAscOrder() {
	q := make(url.Values)
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		}).
		Return([]storage.Log{testLog1, testLog2, testLog3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 3)
	s.Require().EqualValues(1, logs[0].Id)
	s.Require().EqualValues(2, logs[1].Id)
	s.Require().EqualValues(3, logs[2].Id)
}

// TestListDescOrder tests list with descending sort order (default)
func (s *LogHandlerTestSuite) TestListDescOrder() {
	q := make(url.Values)
	q.Set("sort", "desc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.Log{testLog3, testLog2, testLog1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 3)
	s.Require().EqualValues(3, logs[0].Id)
	s.Require().EqualValues(2, logs[1].Id)
	s.Require().EqualValues(1, logs[2].Id)
}

// TestListWithTxHash tests filtering logs by transaction hash
func (s *LogHandlerTestSuite) TestListWithTxHash() {
	q := make(url.Values)
	q.Set("tx_hash", testTxHash)

	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Tx{Id: 1}, nil).
		Times(1)

	txId := uint64(1)
	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			TxId:   &txId,
		}).
		Return([]storage.Log{testLog1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err = json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 1)
	s.Require().EqualValues(1, logs[0].Id)
}

// TestListWithAddress tests filtering logs by address
func (s *LogHandlerTestSuite) TestListWithAddress() {
	q := make(url.Values)
	q.Set("address", testAddressHash)

	hashBytes, err := pkgTypes.HexFromString(testAddressHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.address.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Address{Id: 1}, nil).
		Times(1)

	addressId := uint64(1)
	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:     10,
			Offset:    0,
			Sort:      sdk.SortOrderDesc,
			AddressId: &addressId,
		}).
		Return([]storage.Log{testLog1, testLog3}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err = json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 2)
	s.Require().EqualValues(1, logs[0].Id)
	s.Require().EqualValues(3, logs[1].Id)
}

// TestListWithHeight tests filtering logs by block height
func (s *LogHandlerTestSuite) TestListWithHeight() {
	q := make(url.Values)
	q.Set("height", "100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	height := uint64(100)
	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Height: &height,
		}).
		Return([]storage.Log{testLog1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 1)
	s.Require().EqualValues(100, logs[0].Height)
}

// TestListWithTimeRange tests filtering logs by time range
func (s *LogHandlerTestSuite) TestListWithTimeRange() {
	q := make(url.Values)
	q.Set("from", "1690855260")
	q.Set("to", "1690941660")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.LogListFilter) ([]storage.Log, error) {
			s.Require().Equal(10, filter.Limit)
			s.Require().Equal(0, filter.Offset)
			s.Require().Equal(sdk.SortOrderDesc, filter.Sort)
			s.Require().False(filter.TimeFrom.IsZero())
			s.Require().False(filter.TimeTo.IsZero())
			return []storage.Log{testLog1, testLog2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 2)
}

// TestListEmptyResult tests list when no logs are found
func (s *LogHandlerTestSuite) TestListEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.Log{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 0)
}

// TestListInvalidLimit tests list with invalid limit parameter
func (s *LogHandlerTestSuite) TestListInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.Log{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestListMaxLimit tests list with limit exceeding maximum
func (s *LogHandlerTestSuite) TestListMaxLimit() {
	q := make(url.Values)
	q.Set("limit", "101")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidSort tests list with invalid sort parameter
func (s *LogHandlerTestSuite) TestListInvalidSort() {
	q := make(url.Values)
	q.Set("sort", "invalid")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListNegativeOffset tests list with negative offset
func (s *LogHandlerTestSuite) TestListNegativeOffset() {
	q := make(url.Values)
	q.Set("offset", "-1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidTxHash tests handling of invalid transaction hash
func (s *LogHandlerTestSuite) TestListInvalidTxHash() {
	q := make(url.Values)
	q.Set("tx_hash", "invalid_hash")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidAddress tests handling of invalid address
func (s *LogHandlerTestSuite) TestListInvalidAddress() {
	q := make(url.Values)
	q.Set("address", "invalid_address")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListTxNotFound tests when transaction is not found
func (s *LogHandlerTestSuite) TestListTxNotFound() {
	q := make(url.Values)
	q.Set("tx_hash", testTxHash)

	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Tx{}, sql.ErrNoRows).
		Times(1)

	s.tx.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestListAddressNotFound tests when address is not found
func (s *LogHandlerTestSuite) TestListAddressNotFound() {
	q := make(url.Values)
	q.Set("address", testAddressHash)

	hashBytes, err := pkgTypes.HexFromString(testAddressHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.address.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Address{}, sql.ErrNoRows).
		Times(1)

	s.address.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestListWithInvalidTxHashLength tests handling of invalid hash length
func (s *LogHandlerTestSuite) TestListWithInvalidTxHashLength() {
	q := make(url.Values)
	q.Set("tx_hash", "0x01")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListWithTxHashWithoutPrefix tests handling of hash without 0x prefix
func (s *LogHandlerTestSuite) TestListWithTxHashWithoutPrefix() {
	q := make(url.Values)
	q.Set("tx_hash", "010203")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListWithInvalidAddressLength tests handling of invalid address length
func (s *LogHandlerTestSuite) TestListWithInvalidAddressLength() {
	q := make(url.Values)
	q.Set("address", "0x01")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListWithMultipleFilters tests list with multiple filters combined
func (s *LogHandlerTestSuite) TestListWithMultipleFilters() {
	q := make(url.Values)
	q.Set("height", "100")
	q.Set("limit", "5")
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/log")

	height := uint64(100)
	s.log.EXPECT().
		Filter(gomock.Any(), storage.LogListFilter{
			Limit:  5,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
			Height: &height,
		}).
		Return([]storage.Log{testLog1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var logs []responses.Log
	err := json.NewDecoder(rec.Body).Decode(&logs)
	s.Require().NoError(err)
	s.Require().Len(logs, 1)
	s.Require().EqualValues(100, logs[0].Height)
}

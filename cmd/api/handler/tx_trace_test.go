package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	testTrace1 = &storage.Trace{
		Id:           1,
		Height:       100,
		Time:         testTime,
		TxId:         1,
		From:         1,
		To:           uint64Ptr(2),
		GasLimit:     decimal.NewFromInt(21000),
		Amount:       decimalPtr(decimal.NewFromInt(1000000000000000000)),
		Input:        []byte{},
		TxPosition:   0,
		TraceAddress: []uint64{0},
		Type:         types.Call,
		GasUsed:      decimal.NewFromInt(21000),
		Output:       []byte{},
		Subtraces:    0,
		FromAddress:  testFromAddress,
		ToAddress:    &testToAddress,
		Tx: storage.Tx{
			Hash: pkgTypes.Hex{
				0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
	}

	testTrace2 = &storage.Trace{
		Id:           2,
		Height:       100,
		Time:         testTime,
		TxId:         1,
		From:         2,
		To:           nil,
		GasLimit:     decimal.NewFromInt(100000),
		Amount:       nil,
		Input:        []byte{0x60, 0x60, 0x60},
		TxPosition:   0,
		TraceAddress: []uint64{0, 1},
		Type:         types.Create,
		GasUsed:      decimal.NewFromInt(95000),
		Output:       []byte{0x60, 0x80},
		ContractId:   uint64Ptr(1),
		Subtraces:    0,
		FromAddress:  testToAddress,
		Contract:     &testContract,
		Tx: storage.Tx{
			Hash: pkgTypes.Hex{
				0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
	}
)

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// TxTraceHandlerTestSuite -
type TxTraceHandlerTestSuite struct {
	suite.Suite
	tx      *mock.MockITx
	trace   *mock.MockITrace
	address *mock.MockIAddress
	echo    *echo.Echo
	handler *TxHandler
	ctrl    *gomock.Controller
}

// SetupSuite -
func (s *TxTraceHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.tx = mock.NewMockITx(s.ctrl)
	s.trace = mock.NewMockITrace(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.handler = NewTxHandler(s.tx, s.trace, s.address, testIndexerName)
}

// TearDownSuite -
func (s *TxTraceHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
}

func TestSuiteTxTraceHandler_Run(t *testing.T) {
	suite.Run(t, new(TxTraceHandlerTestSuite))
}

// TestTracesSuccess tests successful retrieval of transaction traces
func (s *TxTraceHandlerTestSuite) TestTracesSuccess() {
	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?tx_hash="+testTxHash, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	testTraces := []*storage.Trace{
		testTrace1,
		testTrace2,
	}

	testTx := storage.Tx{
		Id:   1,
		Hash: hashBytes,
	}

	txId := uint64(1)
	expectedFilter := storage.TraceListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
		TxId:   &txId,
		Type:   []types.TraceType{},
	}

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTx, nil).
		Times(1)

	s.trace.EXPECT().
		Filter(gomock.Any(), expectedFilter).
		Return(testTraces, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err = json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 2)

	s.Require().EqualValues(100, traces[0].Height)
	s.Require().Equal(testTxHash, traces[0].TxHash)
	s.Require().Equal("0x1234567890123456789012345678901234567890", traces[0].FromAddress)
	s.Require().NotNil(traces[0].ToAddress)
	s.Require().Equal("0x0987654321098765432109876543210987654321", *traces[0].ToAddress)
	s.Require().Equal("21000", traces[0].GasLimit.String())
	s.Require().Equal("21000", traces[0].GasUsed.String())
	s.Require().Equal("call", traces[0].Type)

	s.Require().EqualValues(100, traces[1].Height)
	s.Require().Equal(testTxHash, traces[1].TxHash)
	s.Require().Equal("0x0987654321098765432109876543210987654321", traces[1].FromAddress)
	s.Require().NotNil(traces[1].Contract)
	s.Require().Equal("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd", *traces[1].Contract)
	s.Require().Equal("create", traces[1].Type)
}

// TestTracesWithLimit tests retrieval with custom limit
func (s *TxTraceHandlerTestSuite) TestTracesWithLimit() {
	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?tx_hash="+testTxHash+"&limit=5", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	testTx := storage.Tx{
		Id:   1,
		Hash: hashBytes,
	}

	txId := uint64(1)
	expectedFilter := storage.TraceListFilter{
		Limit:  5,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
		TxId:   &txId,
		Type:   []types.TraceType{},
	}

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTx, nil).
		Times(1)

	s.trace.EXPECT().
		Filter(gomock.Any(), expectedFilter).
		Return([]*storage.Trace{testTrace1}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err = json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 1)
}

// TestTracesWithOffset tests retrieval with offset
func (s *TxTraceHandlerTestSuite) TestTracesWithOffset() {
	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?tx_hash="+testTxHash+"&offset=2", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	testTx := storage.Tx{
		Id:   1,
		Hash: hashBytes,
	}

	txId := uint64(1)
	expectedFilter := storage.TraceListFilter{
		Limit:  10,
		Offset: 2,
		Sort:   sdk.SortOrderDesc,
		TxId:   &txId,
		Type:   []types.TraceType{},
	}

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTx, nil).
		Times(1)

	s.trace.EXPECT().
		Filter(gomock.Any(), expectedFilter).
		Return([]*storage.Trace{testTrace2}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err = json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 1)
}

// TestTracesWithSortAsc tests retrieval with ascending sort
func (s *TxTraceHandlerTestSuite) TestTracesWithSortAsc() {
	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?tx_hash="+testTxHash+"&sort=asc", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	testTx := storage.Tx{
		Id:   1,
		Hash: hashBytes,
	}

	txId := uint64(1)
	expectedFilter := storage.TraceListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderAsc,
		TxId:   &txId,
		Type:   []types.TraceType{},
	}

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTx, nil).
		Times(1)

	s.trace.EXPECT().
		Filter(gomock.Any(), expectedFilter).
		Return([]*storage.Trace{testTrace1, testTrace2}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err = json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 2)
}

// TestTracesEmptyResult tests when no traces found
func (s *TxTraceHandlerTestSuite) TestTracesEmptyResult() {
	hashBytes, err := pkgTypes.HexFromString(testTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?tx_hash="+testTxHash, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/traces")

	testTx := storage.Tx{
		Id:   1,
		Hash: hashBytes,
	}

	txId := uint64(1)
	expectedFilter := storage.TraceListFilter{
		Limit:  10,
		Offset: 0,
		Sort:   sdk.SortOrderDesc,
		TxId:   &txId,
		Type:   []types.TraceType{},
	}

	s.tx.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testTx, nil).
		Times(1)

	s.trace.EXPECT().
		Filter(gomock.Any(), expectedFilter).
		Return([]*storage.Trace{}, nil).
		Times(1)

	s.Require().NoError(s.handler.Traces(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var traces []responses.Trace
	err = json.NewDecoder(rec.Body).Decode(&traces)
	s.Require().NoError(err)
	s.Require().Len(traces, 0)
}

// TestTracesInvalidHash tests handling of invalid transaction hash
func (s *TxTraceHandlerTestSuite) TestTracesInvalidHash() {
	req := httptest.NewRequest(http.MethodGet, "/?tx_hash=invalid_hash", nil)
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

// TestTracesInvalidLimit tests handling of invalid limit
func (s *TxTraceHandlerTestSuite) TestTracesInvalidLimit() {
	req := httptest.NewRequest(http.MethodGet, "/?limit=101", nil)
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

// TestTracesInvalidSort tests handling of invalid sort parameter
func (s *TxTraceHandlerTestSuite) TestTracesInvalidSort() {
	req := httptest.NewRequest(http.MethodGet, "/?sort=invalid", nil)
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

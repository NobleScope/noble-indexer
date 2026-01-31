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
	testUserOpHash1 = pkgTypes.MustDecodeHex("0xaabbccddee0011223344556677889900aabbccddee0011223344556677889900")
	testUserOpHash2 = pkgTypes.MustDecodeHex("0x1122334455667788990011223344556677889900aabbccddeeff001122334455")

	testPaymasterAddress = storage.Address{
		Id:   3,
		Hash: testAddressHex3,
	}

	testUserOp1 = storage.ERC4337UserOp{
		Id:                 1,
		Height:             100,
		Time:               testTime,
		TxId:               1,
		Hash:               testUserOpHash1,
		SenderId:           1,
		PaymasterId:        ptrUint64(3),
		BundlerId:          2,
		Nonce:              decimal.NewFromInt(0),
		Success:            true,
		ActualGasCost:      decimal.NewFromInt(100000),
		ActualGasUsed:      decimal.NewFromInt(50000),
		InitCode:           pkgTypes.Hex{},
		CallData:           pkgTypes.Hex{0xAB, 0xCD, 0xEF},
		AccountGasLimits:   pkgTypes.Hex{0x00, 0x00},
		PreVerificationGas: decimal.NewFromInt(21000),
		GasFees:            pkgTypes.Hex{0x00, 0x00},
		PaymasterAndData:   pkgTypes.Hex{},
		Signature:          pkgTypes.Hex{0x12, 0x34},
		Tx: storage.Tx{
			Id:   1,
			Hash: pkgTypes.Hex{0x01, 0x02, 0x03},
		},
		Sender: storage.Address{
			Id:   1,
			Hash: testAddressHex1,
		},
		Paymaster: &testPaymasterAddress,
		Bundler: storage.Address{
			Id:   2,
			Hash: testAddressHex2,
		},
	}

	testUserOp2 = storage.ERC4337UserOp{
		Id:                 2,
		Height:             200,
		Time:               testTime,
		TxId:               4,
		Hash:               testUserOpHash2,
		SenderId:           2,
		BundlerId:          3,
		Nonce:              decimal.NewFromInt(1),
		Success:            false,
		ActualGasCost:      decimal.NewFromInt(200000),
		ActualGasUsed:      decimal.NewFromInt(80000),
		InitCode:           pkgTypes.Hex{},
		CallData:           pkgTypes.Hex{0xDE, 0xAD},
		AccountGasLimits:   pkgTypes.Hex{0x00, 0x00},
		PreVerificationGas: decimal.NewFromInt(21000),
		GasFees:            pkgTypes.Hex{0x00, 0x00},
		PaymasterAndData:   pkgTypes.Hex{},
		Signature:          pkgTypes.Hex{0xFE, 0xDC},
		Tx: storage.Tx{
			Id:   4,
			Hash: pkgTypes.Hex{0x04, 0x05, 0x06},
		},
		Sender: storage.Address{
			Id:   2,
			Hash: testAddressHex2,
		},
		Bundler: storage.Address{
			Id:   3,
			Hash: testAddressHex3,
		},
	}
)

func ptrUint64(v uint64) *uint64 {
	return &v
}

// UserOpHandlerTestSuite -
type UserOpHandlerTestSuite struct {
	suite.Suite
	userOps *mock.MockIERC4337UserOps
	tx      *mock.MockITx
	address *mock.MockIAddress
	echo    *echo.Echo
	handler *UserOpHandler
	ctrl    *gomock.Controller
}

// SetupSuite -
func (s *UserOpHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.userOps = mock.NewMockIERC4337UserOps(s.ctrl)
	s.tx = mock.NewMockITx(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.handler = NewUserOpHandler(s.userOps, s.tx, s.address)
}

// TearDownSuite -
func (s *UserOpHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteUserOpHandler_Run(t *testing.T) {
	suite.Run(t, new(UserOpHandlerTestSuite))
}

// TestListSuccess tests successful retrieval of user ops with default parameters
func (s *UserOpHandlerTestSuite) TestListSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1, testUserOp2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 2)

	s.Require().EqualValues(1, ops[0].Id)
	s.Require().EqualValues(100, ops[0].Height)
	s.Require().Equal(testTime, ops[0].Time)
	s.Require().True(ops[0].Success)
	s.Require().NotNil(ops[0].Paymaster)

	s.Require().EqualValues(2, ops[1].Id)
	s.Require().False(ops[1].Success)
	s.Require().Nil(ops[1].Paymaster)
}

// TestListWithLimit tests list with custom limit parameter
func (s *UserOpHandlerTestSuite) TestListWithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  5,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
}

// TestListWithOffset tests list with offset parameter
func (s *UserOpHandlerTestSuite) TestListWithOffset() {
	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 5,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{testUserOp2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().EqualValues(2, ops[0].Id)
}

// TestListWithLimitAndOffset tests list with both limit and offset
func (s *UserOpHandlerTestSuite) TestListWithLimitAndOffset() {
	q := make(url.Values)
	q.Set("limit", "2")
	q.Set("offset", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  2,
			Offset: 1,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{testUserOp2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
}

// TestListAscOrder tests list with ascending sort order
func (s *UserOpHandlerTestSuite) TestListAscOrder() {
	q := make(url.Values)
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderAsc,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1, testUserOp2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 2)
	s.Require().EqualValues(1, ops[0].Id)
	s.Require().EqualValues(2, ops[1].Id)
}

// TestListDescOrder tests list with descending sort order (explicit)
func (s *UserOpHandlerTestSuite) TestListDescOrder() {
	q := make(url.Values)
	q.Set("sort", "desc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{testUserOp2, testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 2)
	s.Require().EqualValues(2, ops[0].Id)
	s.Require().EqualValues(1, ops[1].Id)
}

// TestListWithHeight tests filtering user ops by block height
func (s *UserOpHandlerTestSuite) TestListWithHeight() {
	q := make(url.Values)
	q.Set("height", "100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	height := uint64(100)
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			Height: &height,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().EqualValues(100, ops[0].Height)
}

// TestListWithTxHash tests filtering user ops by transaction hash
func (s *UserOpHandlerTestSuite) TestListWithTxHash() {
	q := make(url.Values)
	q.Set("tx", testTxHash.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(storage.Tx{Id: 1}, nil).
		Times(1)

	txId := uint64(1)
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
			TxId:   &txId,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().EqualValues(1, ops[0].Id)
}

// TestListWithBundler tests filtering user ops by bundler address
func (s *UserOpHandlerTestSuite) TestListWithBundler() {
	q := make(url.Values)
	q.Set("bundler", testAddressHex2.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex2).
		Return(storage.Address{Id: 2}, nil).
		Times(1)

	bundlerId := uint64(2)
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:     10,
			Offset:    0,
			Sort:      sdk.SortOrderDesc,
			BundlerId: &bundlerId,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().EqualValues(1, ops[0].Id)
}

// TestListWithPaymaster tests filtering user ops by paymaster address
func (s *UserOpHandlerTestSuite) TestListWithPaymaster() {
	q := make(url.Values)
	q.Set("paymaster", testAddressHex3.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(storage.Address{Id: 3}, nil).
		Times(1)

	paymasterId := uint64(3)
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:       10,
			Offset:      0,
			Sort:        sdk.SortOrderDesc,
			PaymasterId: &paymasterId,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().NotNil(ops[0].Paymaster)
}

// TestListWithSuccessTrue tests filtering user ops by success=true
func (s *UserOpHandlerTestSuite) TestListWithSuccessTrue() {
	q := make(url.Values)
	q.Set("success", "true")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	success := true
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:   10,
			Offset:  0,
			Sort:    sdk.SortOrderDesc,
			Success: &success,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().True(ops[0].Success)
}

// TestListWithSuccessFalse tests filtering user ops by success=false
func (s *UserOpHandlerTestSuite) TestListWithSuccessFalse() {
	q := make(url.Values)
	q.Set("success", "false")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	success := false
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:   10,
			Offset:  0,
			Sort:    sdk.SortOrderDesc,
			Success: &success,
		}).
		Return([]storage.ERC4337UserOp{testUserOp2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().False(ops[0].Success)
}

// TestListWithTimeRange tests filtering user ops by time range
func (s *UserOpHandlerTestSuite) TestListWithTimeRange() {
	q := make(url.Values)
	q.Set("from", "1690855260")
	q.Set("to", "1690941660")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.ERC4337UserOpsListFilter) ([]storage.ERC4337UserOp, error) {
			s.Require().Equal(10, filter.Limit)
			s.Require().Equal(0, filter.Offset)
			s.Require().Equal(sdk.SortOrderDesc, filter.Sort)
			s.Require().False(filter.TimeFrom.IsZero())
			s.Require().False(filter.TimeTo.IsZero())
			return []storage.ERC4337UserOp{testUserOp1}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
}

// TestListEmptyResult tests list when no user ops are found
func (s *UserOpHandlerTestSuite) TestListEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 0)
}

// TestListInvalidLimit tests list with limit=0 defaults to 10
func (s *UserOpHandlerTestSuite) TestListInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestListMaxLimit tests list with limit exceeding maximum
func (s *UserOpHandlerTestSuite) TestListMaxLimit() {
	q := make(url.Values)
	q.Set("limit", "101")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidSort tests list with invalid sort parameter
func (s *UserOpHandlerTestSuite) TestListInvalidSort() {
	q := make(url.Values)
	q.Set("sort", "invalid")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListNegativeOffset tests list with negative offset
func (s *UserOpHandlerTestSuite) TestListNegativeOffset() {
	q := make(url.Values)
	q.Set("offset", "-1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidTxHash tests handling of invalid transaction hash
func (s *UserOpHandlerTestSuite) TestListInvalidTxHash() {
	q := make(url.Values)
	q.Set("tx", "invalid_hash")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListWithInvalidTxHashLength tests handling of short tx hash
func (s *UserOpHandlerTestSuite) TestListWithInvalidTxHashLength() {
	q := make(url.Values)
	q.Set("tx", "0x01")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidBundlerAddress tests handling of invalid bundler address
func (s *UserOpHandlerTestSuite) TestListInvalidBundlerAddress() {
	q := make(url.Values)
	q.Set("bundler", "invalid_address")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListWithInvalidBundlerAddressLength tests handling of short bundler address
func (s *UserOpHandlerTestSuite) TestListWithInvalidBundlerAddressLength() {
	q := make(url.Values)
	q.Set("bundler", "0x01")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidPaymasterAddress tests handling of invalid paymaster address
func (s *UserOpHandlerTestSuite) TestListInvalidPaymasterAddress() {
	q := make(url.Values)
	q.Set("paymaster", "invalid_address")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListTxNotFound tests when transaction is not found
func (s *UserOpHandlerTestSuite) TestListTxNotFound() {
	q := make(url.Values)
	q.Set("tx", testTxHash.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(storage.Tx{}, sql.ErrNoRows).
		Times(1)

	s.tx.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestListBundlerNotFound tests when bundler address is not found
func (s *UserOpHandlerTestSuite) TestListBundlerNotFound() {
	q := make(url.Values)
	q.Set("bundler", testAddressHex1.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{}, sql.ErrNoRows).
		Times(1)

	s.address.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestListPaymasterNotFound tests when paymaster address is not found
func (s *UserOpHandlerTestSuite) TestListPaymasterNotFound() {
	q := make(url.Values)
	q.Set("paymaster", testAddressHex1.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Address{}, sql.ErrNoRows).
		Times(1)

	s.address.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

// TestListFilterError tests internal server error from Filter
func (s *UserOpHandlerTestSuite) TestListFilterError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	testErr := sql.ErrConnDone
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return(nil, testErr).
		Times(1)

	s.userOps.EXPECT().
		IsNoRows(testErr).
		Return(false).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusInternalServerError, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListWithMultipleFilters tests list with multiple filters combined
func (s *UserOpHandlerTestSuite) TestListWithMultipleFilters() {
	q := make(url.Values)
	q.Set("height", "100")
	q.Set("success", "true")
	q.Set("limit", "5")
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	height := uint64(100)
	success := true
	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:   5,
			Offset:  0,
			Sort:    sdk.SortOrderAsc,
			Height:  &height,
			Success: &success,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
	s.Require().EqualValues(100, ops[0].Height)
	s.Require().True(ops[0].Success)
}

// TestListWithAllFilters tests list with all filters combined
func (s *UserOpHandlerTestSuite) TestListWithAllFilters() {
	q := make(url.Values)
	q.Set("height", "100")
	q.Set("tx", testTxHash.Hex())
	q.Set("bundler", testAddressHex2.Hex())
	q.Set("paymaster", testAddressHex3.Hex())
	q.Set("success", "true")
	q.Set("from", "1690855260")
	q.Set("to", "1690941660")
	q.Set("limit", "5")
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.tx.EXPECT().
		ByHash(gomock.Any(), testTxHash).
		Return(storage.Tx{Id: 1}, nil).
		Times(1)

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex2).
		Return(storage.Address{Id: 2}, nil).
		Times(1)

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(storage.Address{Id: 3}, nil).
		Times(1)

	s.userOps.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.ERC4337UserOpsListFilter) ([]storage.ERC4337UserOp, error) {
			s.Require().Equal(5, filter.Limit)
			s.Require().Equal(0, filter.Offset)
			s.Require().Equal(sdk.SortOrderAsc, filter.Sort)
			s.Require().NotNil(filter.Height)
			s.Require().EqualValues(100, *filter.Height)
			s.Require().NotNil(filter.TxId)
			s.Require().EqualValues(1, *filter.TxId)
			s.Require().NotNil(filter.BundlerId)
			s.Require().EqualValues(2, *filter.BundlerId)
			s.Require().NotNil(filter.PaymasterId)
			s.Require().EqualValues(3, *filter.PaymasterId)
			s.Require().NotNil(filter.Success)
			s.Require().True(*filter.Success)
			s.Require().False(filter.TimeFrom.IsZero())
			s.Require().False(filter.TimeTo.IsZero())
			return []storage.ERC4337UserOp{testUserOp1}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)
}

// TestListResponseFields tests that all response fields are correctly mapped
func (s *UserOpHandlerTestSuite) TestListResponseFields() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/user_ops")

	s.userOps.EXPECT().
		Filter(gomock.Any(), storage.ERC4337UserOpsListFilter{
			Limit:  10,
			Offset: 0,
			Sort:   sdk.SortOrderDesc,
		}).
		Return([]storage.ERC4337UserOp{testUserOp1}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var ops []responses.UserOp
	err := json.NewDecoder(rec.Body).Decode(&ops)
	s.Require().NoError(err)
	s.Require().Len(ops, 1)

	op := ops[0]
	s.Require().EqualValues(1, op.Id)
	s.Require().EqualValues(100, op.Height)
	s.Require().Equal(testTime, op.Time)
	s.Require().Equal(testUserOp1.Tx.Hash.Hex(), op.TxHash)
	s.Require().Equal(testUserOp1.Hash.Hex(), op.Hash)
	s.Require().Equal(testUserOp1.Sender.Hash.Hex(), op.Sender)
	s.Require().NotNil(op.Paymaster)
	s.Require().Equal(testUserOp1.Paymaster.Hash.Hex(), *op.Paymaster)
	s.Require().Equal(testUserOp1.Bundler.Hash.Hex(), op.Bundler)
	s.Require().True(op.Nonce.Equal(decimal.NewFromInt(0)))
	s.Require().True(op.Success)
	s.Require().True(op.ActualGasCost.Equal(decimal.NewFromInt(100000)))
	s.Require().True(op.ActualGasUsed.Equal(decimal.NewFromInt(50000)))
	s.Require().True(op.PreVerificationGas.Equal(decimal.NewFromInt(21000)))
	s.Require().NotEmpty(op.Signature)
}

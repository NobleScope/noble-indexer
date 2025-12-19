package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func uint64Ptr(v uint64) *uint64 {
	return &v
}

var (
	testFromAddress = storage.Address{
		Id:         1,
		Height:     100,
		LastHeight: 100,
		Hash:       testAddressHex1,
		IsContract: false,
	}

	testToAddress = storage.Address{
		Id:         2,
		Height:     100,
		LastHeight: 100,
		Hash:       testAddressHex2,
		IsContract: false,
	}

	testContract = storage.Contract{
		Id: 1,
		Address: storage.Address{
			Hash: testAddressHex3,
		},
		Code:     []byte{0x60, 0x60, 0x60},
		Verified: true,
		TxId:     uint64Ptr(1),
	}

	testTxWithToAddress = storage.Tx{
		Id:                1,
		Height:            100,
		Time:              testTime,
		Gas:               decimal.NewFromInt(21000),
		GasPrice:          decimal.NewFromInt(1000000),
		Hash:              pkgTypes.Hex{0x01, 0x02, 0x03},
		Nonce:             1,
		Index:             0,
		Amount:            decimal.NewFromInt(1000000000000000000),
		Type:              types.TxTypeDynamicFee,
		Input:             pkgTypes.Hex{},
		CumulativeGasUsed: decimal.NewFromInt(21000),
		EffectiveGasPrice: decimal.NewFromInt(1000000),
		FromAddressId:     1,
		ToAddressId:       uint64Ptr(2),
		Fee:               decimal.NewFromInt(21000000000),
		GasUsed:           decimal.NewFromInt(21000),
		Status:            types.TxStatusSuccess,
		LogsBloom:         pkgTypes.Hex{0x00},
		FromAddress:       testFromAddress,
		ToAddress:         &testToAddress,
	}

	testTxContractCreation = storage.Tx{
		Id:                2,
		Height:            100,
		Time:              testTime,
		Gas:               decimal.NewFromInt(100000),
		GasPrice:          decimal.NewFromInt(2000000),
		Hash:              pkgTypes.Hex{0x04, 0x05, 0x06},
		Nonce:             2,
		Index:             1,
		Amount:            decimal.Zero,
		Type:              types.TxTypeDynamicFee,
		Input:             pkgTypes.Hex{0x60, 0x60, 0x60},
		CumulativeGasUsed: decimal.NewFromInt(121000),
		EffectiveGasPrice: decimal.NewFromInt(2000000),
		FromAddressId:     1,
		ToAddressId:       nil,
		Fee:               decimal.NewFromInt(200000000000),
		GasUsed:           decimal.NewFromInt(100000),
		Status:            types.TxStatusSuccess,
		LogsBloom:         pkgTypes.Hex{0x00},
		FromAddress:       testFromAddress,
		ToAddress:         nil,
	}

	testTxContractCall = storage.Tx{
		Id:                3,
		Height:            100,
		Time:              testTime,
		Gas:               decimal.NewFromInt(50000),
		GasPrice:          decimal.NewFromInt(1500000),
		Hash:              pkgTypes.Hex{0x07, 0x08, 0x09},
		Nonce:             3,
		Index:             2,
		Amount:            decimal.NewFromInt(100000000),
		Type:              types.TxTypeDynamicFee,
		Input:             pkgTypes.Hex{0xa9, 0x05, 0x9c, 0xbb},
		CumulativeGasUsed: decimal.NewFromInt(171000),
		EffectiveGasPrice: decimal.NewFromInt(1500000),
		FromAddressId:     1,
		ToAddressId:       uint64Ptr(2),
		Fee:               decimal.NewFromInt(75000000000),
		GasUsed:           decimal.NewFromInt(50000),
		Status:            types.TxStatusSuccess,
		LogsBloom:         pkgTypes.Hex{0x00},
		FromAddress:       testFromAddress,
		ToAddress:         &testToAddress,
	}
)

// TxTestSuite -
type TxTestSuite struct {
	suite.Suite
	txs     *mock.MockITx
	echo    *echo.Echo
	handler *BlockHandler
	ctrl    *gomock.Controller
}

// SetupTest -
func (s *TxTestSuite) SetupTest() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.txs = mock.NewMockITx(s.ctrl)
	s.handler = NewBlockHandler(nil, nil, s.txs, nil, testIndexerName)
}

// TearDownTest -
func (s *TxTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestSuiteTx_Run(t *testing.T) {
	suite.Run(t, new(TxTestSuite))
}

// TestTxsList
func (s *TxTestSuite) TestTxsList() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), 10, 0, sdk.SortOrderAsc).
		Return([]*storage.Tx{
			&testTxWithToAddress,
			&testTxContractCreation,
			&testTxContractCall,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
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
	s.Require().Equal(testAddressHex1.Hex(), txs[1].FromAddress)
	s.Require().Nil(txs[1].ToAddress)

	s.Require().EqualValues(100, txs[2].Height)
	s.Require().Equal("0x070809", txs[2].Hash)
	s.Require().EqualValues(2, txs[2].Index)
	s.Require().Equal(testAddressHex1.Hex(), txs[2].FromAddress)
	s.Require().NotNil(txs[2].ToAddress)
	s.Require().Equal(testAddressHex2.Hex(), *txs[2].ToAddress)
}

// TestTxsListWithLimit
func (s *TxTestSuite) TestTxsListWithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), 5, 0, sdk.SortOrderAsc).
		Return([]*storage.Tx{
			&testTxWithToAddress,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 1)
}

// TestTxsListWithOffset
func (s *TxTestSuite) TestTxsListWithOffset() {
	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), 10, 5, sdk.SortOrderAsc).
		Return([]*storage.Tx{
			&testTxContractCall,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 1)
	s.Require().EqualValues(2, txs[0].Index)
}

// TestTxsListWithLimitAndOffset
func (s *TxTestSuite) TestTxsListWithLimitAndOffset() {
	q := make(url.Values)
	q.Set("limit", "2")
	q.Set("offset", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), 2, 1, sdk.SortOrderAsc).
		Return([]*storage.Tx{
			&testTxContractCreation,
			&testTxContractCall,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 2)
	s.Require().EqualValues(1, txs[0].Index)
	s.Require().EqualValues(2, txs[1].Index)
}

// TestTxsListDescOrder
func (s *TxTestSuite) TestTxsListDescOrder() {
	q := make(url.Values)
	q.Set("sort", "desc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), 10, 0, sdk.SortOrderDesc).
		Return([]*storage.Tx{
			&testTxContractCall,
			&testTxContractCreation,
			&testTxWithToAddress,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 3)
	s.Require().EqualValues(2, txs[0].Index)
	s.Require().EqualValues(1, txs[1].Index)
	s.Require().EqualValues(0, txs[2].Index)
}

// TestTxsListEmptyResult
func (s *TxTestSuite) TestTxsListEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("999")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(999), 10, 0, sdk.SortOrderAsc).
		Return([]*storage.Tx{}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 0)
}

// TestTxsListInvalidHeight
func (s *TxTestSuite) TestTxsListInvalidHeight() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("invalid")

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Contains(e.Message, "parsing")
}

// TestTxsListInvalidLimit
func (s *TxTestSuite) TestTxsListInvalidLimit() {
	q := make(url.Values)
	q.Set("limit", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), 10, 0, sdk.SortOrderAsc).
		Return([]*storage.Tx{}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusOK, rec.Code)
}

// TestTxsListMaxLimit
func (s *TxTestSuite) TestTxsListMaxLimit() {
	q := make(url.Values)
	q.Set("limit", "101") // больше максимума 100

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// TestTxsOnlyFromAddress
func (s *TxTestSuite) TestTxsOnlyFromAddress() {
	minimalTx := storage.Tx{
		Id:                4,
		Height:            100,
		Time:              testTime,
		Gas:               decimal.NewFromInt(21000),
		GasPrice:          decimal.NewFromInt(1000000),
		Hash:              pkgTypes.Hex{0x0a, 0x0b, 0x0c},
		Nonce:             4,
		Index:             3,
		Amount:            decimal.Zero,
		Type:              types.TxTypeLegacy,
		Input:             pkgTypes.Hex{},
		CumulativeGasUsed: decimal.NewFromInt(192000),
		EffectiveGasPrice: decimal.NewFromInt(1000000),
		FromAddressId:     1,
		ToAddressId:       nil,
		Fee:               decimal.NewFromInt(21000000000),
		GasUsed:           decimal.NewFromInt(21000),
		Status:            types.TxStatusRevert,
		LogsBloom:         pkgTypes.Hex{0x00},
		FromAddress:       testFromAddress,
		ToAddress:         nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height/transactions")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.txs.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), 10, 0, sdk.SortOrderAsc).
		Return([]*storage.Tx{&minimalTx}, nil).
		Times(1)

	s.Require().NoError(s.handler.TransactionsList(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var txs []responses.Transaction
	err := json.NewDecoder(rec.Body).Decode(&txs)
	s.Require().NoError(err)
	s.Require().Len(txs, 1)
	s.Require().Equal("0x1234567890123456789012345678901234567890", txs[0].FromAddress)
	s.Require().Nil(txs[0].ToAddress)
	s.Require().Equal("TxStatusRevert", txs[0].Status)
}

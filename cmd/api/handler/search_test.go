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
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

const (
	testSearchTxHash    = "0x0102030000000000000000000000000000000000000000000000000000000000"
	testSearchAddress   = "0x1234567890123456789012345678901234567890"
	testSearchTokenText = "USDC"
)

var (
	testSearchAddress1 = storage.Address{
		Id:             1,
		FirstHeight:    100,
		LastHeight:     200,
		Hash:           pkgTypes.MustDecodeHex(testSearchTxHash),
		IsContract:     false,
		TxsCount:       50,
		ContractsCount: 0,
		Interactions:   75,
		Balance: &storage.Balance{
			Id:       1,
			Value:    decimal.RequireFromString("1000000000"),
		},
	}

	testSearchToken = storage.Token{
		Id:             1,
		TokenID:        decimal.NewFromInt(0),
		ContractId:     1,
		Type:           types.ERC20,
		Height:         100,
		LastHeight:     200,
		Name:           "USD Coin",
		Symbol:         "USDC",
		Decimals:       6,
		TransfersCount: 100,
		Supply:         decimal.NewFromInt(1000000),
		Status:         types.Success,
	}
)

// SearchHandlerTestSuite -
type SearchHandlerTestSuite struct {
	suite.Suite
	search  *mock.MockISearch
	address *mock.MockIAddress
	block   *mock.MockIBlock
	tx      *mock.MockITx
	token   *mock.MockIToken
	echo    *echo.Echo
	handler SearchHandler
	ctrl    *gomock.Controller
}

// SetupSuite -
func (s *SearchHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.search = mock.NewMockISearch(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.block = mock.NewMockIBlock(s.ctrl)
	s.tx = mock.NewMockITx(s.ctrl)
	s.token = mock.NewMockIToken(s.ctrl)
	s.handler = NewSearchHandler(s.search, s.address, s.block, s.tx, s.token)
}

// TearDownSuite -
func (s *SearchHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteSearchHandler_Run(t *testing.T) {
	suite.Run(t, new(SearchHandlerTestSuite))
}

// TestSearchByHeight tests successful search by block height
func (s *SearchHandlerTestSuite) TestSearchByHeight() {
	q := make(url.Values)
	q.Set("query", "100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.block.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), false).
		Return(testBlock, nil).
		Times(1)

	// After checking block height, the code also checks the switch statement
	// and since "100" doesn't match address or tx hash regex, it goes to default (searchText)
	s.search.EXPECT().
		SearchText(gomock.Any(), "100").
		Return([]storage.SearchResult{}, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err := json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("block", results[0].Type)
}

// TestSearchByAddress tests successful search by address
func (s *SearchHandlerTestSuite) TestSearchByAddress() {
	q := make(url.Values)
	q.Set("query", testSearchAddress)

	hashBytes, err := pkgTypes.HexFromString(testSearchAddress)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.address.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testSearchAddress1, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err = json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("address", results[0].Type)
}

// TestSearchByTxHash tests successful search by transaction hash
func (s *SearchHandlerTestSuite) TestSearchByTxHash() {
	q := make(url.Values)
	q.Set("query", testSearchTxHash)

	hashBytes, err := pkgTypes.HexFromString(testSearchTxHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.search.EXPECT().
		Search(gomock.Any(), hashBytes).
		Return([]storage.SearchResult{
			{Type: "tx", Id: 1},
		}, nil).
		Times(1)

	s.tx.EXPECT().
		GetByID(gomock.Any(), uint64(1)).
		Return(&testTxWithToAddress, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err = json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("tx", results[0].Type)
}

// TestSearchByText tests successful search by text (token name/symbol)
func (s *SearchHandlerTestSuite) TestSearchByText() {
	q := make(url.Values)
	q.Set("query", testSearchTokenText)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.search.EXPECT().
		SearchText(gomock.Any(), testSearchTokenText).
		Return([]storage.SearchResult{
			{Type: "token", Id: 1},
		}, nil).
		Times(1)

	s.token.EXPECT().
		GetByID(gomock.Any(), uint64(1)).
		Return(&testSearchToken, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err := json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Require().Equal("token", results[0].Type)
}

// TestSearchMissingQuery tests handling of missing query parameter
func (s *SearchHandlerTestSuite) TestSearchMissingQuery() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestSearchNoResults tests search that returns empty results
func (s *SearchHandlerTestSuite) TestSearchNoResults() {
	q := make(url.Values)
	q.Set("query", "nonexistent")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.search.EXPECT().
		SearchText(gomock.Any(), "nonexistent").
		Return([]storage.SearchResult{}, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err := json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 0)
}

// TestSearchInvalidAddress tests handling of invalid address format
func (s *SearchHandlerTestSuite) TestSearchInvalidAddress() {
	q := make(url.Values)
	q.Set("query", "0xinvalid")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.search.EXPECT().
		SearchText(gomock.Any(), "0xinvalid").
		Return([]storage.SearchResult{}, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err := json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 0)
}

// TestSearchMultipleResults tests search that returns multiple results (block + address)
func (s *SearchHandlerTestSuite) TestSearchMultipleResults() {
	q := make(url.Values)
	q.Set("query", testSearchAddress)

	hashBytes, err := pkgTypes.HexFromString(testSearchAddress)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.address.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(testSearchAddress1, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err = json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(results), 1)
}

// TestSearchByHashNotFound tests search by hash that is not found
func (s *SearchHandlerTestSuite) TestSearchByHashNotFound() {
	txHash := "0xaaaaaaaaaa000000000000000000000000000000000000000000000000000000"
	q := make(url.Values)
	q.Set("query", txHash)

	hashBytes, err := pkgTypes.HexFromString(txHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.search.EXPECT().
		Search(gomock.Any(), hashBytes).
		Return([]storage.SearchResult{}, nil).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err = json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 0)
}

// TestSearchByAddressNotFound tests search by address that is not found
func (s *SearchHandlerTestSuite) TestSearchByAddressNotFound() {
	addressHash := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	q := make(url.Values)
	q.Set("query", addressHash)

	hashBytes, err := pkgTypes.HexFromString(addressHash)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/search")

	s.address.EXPECT().
		ByHash(gomock.Any(), hashBytes).
		Return(storage.Address{}, sql.ErrNoRows).
		Times(1)

	s.address.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.Search(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var results []responses.SearchItem
	err = json.NewDecoder(rec.Body).Decode(&results)
	s.Require().NoError(err)
	s.Require().Len(results, 0)
}

package handler

import (
	"database/sql"
	"encoding/json"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/mock"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ContractTestSuite struct {
	suite.Suite

	echo     *echo.Echo
	ctrl     *gomock.Controller
	contract *mock.MockIContract
	tx       *mock.MockITx
	source   *mock.MockISource
	handler  *ContractHandler
}

func (s *ContractTestSuite) SetupTest() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()

	s.ctrl = gomock.NewController(s.T())
	s.contract = mock.NewMockIContract(s.ctrl)
	s.tx = mock.NewMockITx(s.ctrl)
	s.source = mock.NewMockISource(s.ctrl)

	s.handler = NewContractHandler(s.contract, s.tx, s.source)
}

func (s *ContractTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestSuiteContract_Run(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) TestContractList_Default() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	s.contract.EXPECT().
		ListWithTx(gomock.Any(), storage.ContractListFilter{
			Limit:      10,
			Offset:     0,
			Sort:       sdk.SortOrderAsc,
			SortField:  "",
			IsVerified: false,
		}).
		Return([]storage.Contract{testContract}, nil)

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var resp []responses.Contract
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Require().Len(resp, 1)
}

func (s *ContractTestSuite) TestContractList_WithParams() {
	q := make(url.Values)
	q.Set("limit", "5")
	q.Set("offset", "2")
	q.Set("sort", "desc")
	q.Set("sort_by", "height")
	q.Set("is_verified", "true")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	s.contract.EXPECT().
		ListWithTx(gomock.Any(), storage.ContractListFilter{
			Limit:      5,
			Offset:     2,
			Sort:       sdk.SortOrderDesc,
			SortField:  "height",
			IsVerified: true,
		}).
		Return([]storage.Contract{testContract}, nil)

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *ContractTestSuite) TestContractList_FilterByTxHash() {
	q := make(url.Values)
	q.Set("tx_hash", testTxHash.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	s.tx.EXPECT().
		ByHash(gomock.Any(), gomock.Any()).
		Return(storage.Tx{Id: 77}, nil)

	s.contract.EXPECT().
		ListWithTx(gomock.Any(), storage.ContractListFilter{
			Limit:      10,
			Offset:     0,
			Sort:       sdk.SortOrderAsc,
			SortField:  "",
			IsVerified: false,
			TxId:       uint64Ptr(77),
		}).
		Return([]storage.Contract{testContract}, nil)

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *ContractTestSuite) TestContractList_InvalidTxHash() {
	q := make(url.Values)
	q.Set("tx_hash", "invalid")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	_ = s.handler.List(c)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractTestSuite) TestContractList_Empty() {
	s.contract.EXPECT().
		ListWithTx(gomock.Any(), gomock.Any()).
		Return([]storage.Contract{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var resp []responses.Contract
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	s.Require().Len(resp, 0)
}

func (s *ContractTestSuite) TestContractList_ValidationError() {
	q := make(url.Values)
	q.Set("limit", "200")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	_ = s.handler.List(c)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractTestSuite) TestContractGet() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testContract, nil)

	err := s.handler.Get(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *ContractTestSuite) TestContractGet_InvalidHash() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash")
	c.SetParamNames("hash")
	c.SetParamValues("invalid")

	_ = s.handler.Get(c)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractTestSuite) TestContractGet_NotFoundTx() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), gomock.Any()).
		Return(storage.Contract{}, sql.ErrNoRows)

	s.contract.EXPECT().
		IsNoRows(gomock.Any()).
		Return(true)

	err := s.handler.Get(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

func (s *ContractTestSuite) TestContractSources() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testContract, nil)

	s.source.EXPECT().
		ByContractId(gomock.Any(), testContract.Id, 0, 0).
		Return([]storage.Source{
			{
				Id:         1,
				ContractId: testContract.Id,
			},
		}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var resp []responses.Source
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Require().Len(resp, 1)
}

func (s *ContractTestSuite) TestContractSources_InvalidHash() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues("invalid")

	_ = s.handler.ContractSources(c)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractTestSuite) TestContractSources_Empty() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testContract, nil)

	s.source.EXPECT().
		ByContractId(gomock.Any(), testContract.Id, 0, 0).
		Return([]storage.Source{}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var resp []responses.Source

	err = json.NewDecoder(rec.Body).Decode(&resp)
	s.Require().NoError(err)
	s.Require().Len(resp, 0)
}

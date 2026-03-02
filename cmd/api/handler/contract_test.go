package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/cmd/api/helpers"
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/mock"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ContractTestSuite struct {
	suite.Suite

	echo     *echo.Echo
	ctrl     *gomock.Controller
	contract *mock.MockIContract
	address  *mock.MockIAddress
	tx       *mock.MockITx
	source   *mock.MockISource
	handler  *ContractHandler
}

func (s *ContractTestSuite) SetupTest() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()

	s.ctrl = gomock.NewController(s.T())
	s.contract = mock.NewMockIContract(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.tx = mock.NewMockITx(s.ctrl)
	s.source = mock.NewMockISource(s.ctrl)

	s.handler = NewContractHandler(s.contract, s.address, s.tx, s.source)
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

	var body struct {
		Result []responses.Contract `json:"result"`
		Cursor string               `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	resp := body.Result
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
		ByHash(gomock.Any(), gomock.Any(), false).
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

	var body struct {
		Result []responses.Contract `json:"result"`
		Cursor string               `json:"cursor"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	resp := body.Result
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
		Filter(gomock.Any(), storage.SourceListFilter{
			ContractId: testContract.Id,
			Limit:      0,
			Offset:     0,
			Sort:       sdk.SortOrderAsc,
		}).
		Return([]storage.Source{
			{
				Id:         1,
				ContractId: testContract.Id,
			},
		}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Source `json:"result"`
		Cursor string             `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	resp := body.Result
	s.Require().Len(resp, 1)
}

func (s *ContractTestSuite) TestContractSources_WithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testContract, nil)

	s.source.EXPECT().
		Filter(gomock.Any(), storage.SourceListFilter{
			ContractId: testContract.Id,
			Limit:      5,
			Offset:     0,
			Sort:       sdk.SortOrderAsc,
		}).
		Return([]storage.Source{
			{Id: 1, ContractId: testContract.Id, Name: "a.sol"},
			{Id: 2, ContractId: testContract.Id, Name: "b.sol"},
		}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Source `json:"result"`
		Cursor string             `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	s.Require().Len(body.Result, 2)
	s.Require().NotEmpty(body.Cursor)
	s.Equal("a.sol", body.Result[0].Name)
	s.Equal("b.sol", body.Result[1].Name)
}

func (s *ContractTestSuite) TestContractSources_WithCursor() {
	q := make(url.Values)
	q.Set("cursor", "AAAAAAAAAAU") // EncodeIDCursor(5)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testContract, nil)

	s.source.EXPECT().
		Filter(gomock.Any(), storage.SourceListFilter{
			ContractId: testContract.Id,
			Limit:      0,
			Offset:     0,
			Sort:       sdk.SortOrderAsc,
			CursorID:   5,
		}).
		Return([]storage.Source{
			{Id: 6, ContractId: testContract.Id, Name: "f.sol"},
			{Id: 7, ContractId: testContract.Id, Name: "g.sol"},
		}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Source `json:"result"`
		Cursor string             `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	s.Require().Len(body.Result, 2)
	s.Require().NotEmpty(body.Cursor)
	s.EqualValues(6, body.Result[0].Id)
	s.EqualValues(7, body.Result[1].Id)
}

func (s *ContractTestSuite) TestContractSources_InvalidCursor() {
	q := make(url.Values)
	q.Set("cursor", "not-valid-base64!!!")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testContract, nil)

	_ = s.handler.ContractSources(c)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractTestSuite) TestContractSources_CursorResponseFormat() {
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
		Filter(gomock.Any(), gomock.Any()).
		Return([]storage.Source{}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Source `json:"result"`
		Cursor string             `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	s.Require().Len(body.Result, 0)
	s.Empty(body.Cursor)
}

func (s *ContractTestSuite) TestContractSources_ContractNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(storage.Contract{}, sql.ErrNoRows)

	s.contract.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

func (s *ContractTestSuite) TestContractSources_WithLimitAndOffset() {
	q := make(url.Values)
	q.Set("limit", "2")
	q.Set("offset", "3")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/sources")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testContract, nil)

	s.source.EXPECT().
		Filter(gomock.Any(), storage.SourceListFilter{
			ContractId: testContract.Id,
			Limit:      2,
			Offset:     3,
			Sort:       sdk.SortOrderAsc,
		}).
		Return([]storage.Source{
			{Id: 4, ContractId: testContract.Id},
		}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Source `json:"result"`
		Cursor string             `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	s.Require().Len(body.Result, 1)
	s.EqualValues(4, body.Result[0].Id)
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
		Filter(gomock.Any(), storage.SourceListFilter{
			ContractId: testContract.Id,
			Limit:      0,
			Offset:     0,
			Sort:       sdk.SortOrderAsc,
		}).
		Return([]storage.Source{}, nil)

	err := s.handler.ContractSources(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Source `json:"result"`
		Cursor string             `json:"cursor"`
	}
	err = json.NewDecoder(rec.Body).Decode(&body)
	s.Require().NoError(err)
	resp := body.Result
	s.Require().Len(resp, 0)
}

func (s *ContractTestSuite) TestContractListByDeployer() {
	q := make(url.Values)
	q.Set("deployer", testAddressHex1.Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contracts")

	s.address.EXPECT().
		ByHash(gomock.Any(), testAddressHex1).
		Return(testAddress1, nil).
		Times(1)

	s.contract.EXPECT().
		ListWithTx(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, clf storage.ContractListFilter) ([]storage.Contract, error) {
			s.Equal(uint64(1), *clf.DeployerId)
			return nil, nil
		}).
		Return([]storage.Contract{testContract}, nil).
		Times(1)

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Contract `json:"result"`
		Cursor string               `json:"cursor"`
	}
	err = json.NewDecoder(rec.Body).Decode(&body)
	s.Require().NoError(err)
	resp := body.Result
	s.Require().Len(resp, 1)
}

func (s *ContractTestSuite) TestContractCode() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract/:hash/code")
	c.SetParamNames("hash")
	c.SetParamValues(testAddressHex1.Hex())

	s.contract.EXPECT().
		Code(gomock.Any(), testAddressHex1).
		Return(pkgTypes.MustDecodeHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"), []byte(`{"abi": "test"}`), nil).
		Times(1)

	err := s.handler.GetCode(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var resp responses.ContractCode
	err = json.NewDecoder(rec.Body).Decode(&resp)
	s.Require().NoError(err)
	s.Equal("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", resp.Code)
	s.NotEmpty(resp.ABI)
}

// TestContractList_WithCursor tests cursor-based pagination for contract list
func (s *ContractTestSuite) TestContractList_WithCursor() {
	q := make(url.Values)
	q.Set("cursor", helpers.EncodeIDCursor(1))

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	s.contract.EXPECT().
		ListWithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.ContractListFilter) ([]storage.Contract, error) {
			s.Require().EqualValues(1, filter.CursorID)
			return []storage.Contract{testContract}, nil
		}).
		Times(1)

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Contract `json:"result"`
		Cursor string               `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	s.Require().Len(body.Result, 1)
	s.Require().NotEmpty(body.Cursor)

	decodedID, err := helpers.DecodeIDCursor(body.Cursor)
	s.Require().NoError(err)
	s.Require().EqualValues(testContract.Id, decodedID)
}

// TestContractList_InvalidCursor tests handling of invalid cursor for contract list
func (s *ContractTestSuite) TestContractList_InvalidCursor() {
	q := make(url.Values)
	q.Set("cursor", "not-valid-base64!!!")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// TestContractList_NoCursorOnNonIdSort tests that cursor is empty when sorting by height
func (s *ContractTestSuite) TestContractList_NoCursorOnNonIdSort() {
	q := make(url.Values)
	q.Set("sort_by", "height")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/contract")

	s.contract.EXPECT().
		ListWithTx(gomock.Any(), gomock.Any()).
		Return([]storage.Contract{testContract}, nil)

	err := s.handler.List(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.Contract `json:"result"`
		Cursor string               `json:"cursor"`
	}
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	s.Require().Len(body.Result, 1)
	s.Require().Empty(body.Cursor)
}

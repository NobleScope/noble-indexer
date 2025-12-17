package handler

import (
	"context"
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
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var (
	testProxyContractAddress = storage.Address{
		Id:         10,
		Height:     100,
		LastHeight: 100,
		Address:    "0x1111111111111111111111111111111111111111",
		IsContract: true,
	}

	testImplementationAddress = storage.Address{
		Id:         11,
		Height:     100,
		LastHeight: 100,
		Address:    "0x2222222222222222222222222222222222222222",
		IsContract: true,
	}

	testProxyContract1 = storage.ProxyContract{
		Id:                10,
		Height:            100,
		Type:              types.EIP1967,
		Status:            types.Resolved,
		ResolvingAttempts: 1,
		ImplementationID:  uint64Ptr(11),
		Contract: storage.Contract{
			Id:      10,
			Height:  100,
			Address: testProxyContractAddress,
		},
		Implementation: &storage.Contract{
			Id:      11,
			Height:  100,
			Address: testImplementationAddress,
		},
	}

	testProxyContract2 = storage.ProxyContract{
		Id:                12,
		Height:            101,
		Type:              types.EIP1967,
		Status:            types.New,
		ResolvingAttempts: 0,
		ImplementationID:  nil,
		Contract: storage.Contract{
			Id:     12,
			Height: 101,
			Address: storage.Address{
				Id:         12,
				Height:     101,
				LastHeight: 101,
				Address:    "0x3333333333333333333333333333333333333333",
				IsContract: true,
			},
		},
		Implementation: nil,
	}

	testProxyContract3 = storage.ProxyContract{
		Id:                13,
		Height:            102,
		Type:              types.EIP7702,
		Status:            types.Error,
		ResolvingAttempts: 3,
		ImplementationID:  nil,
		Contract: storage.Contract{
			Id:     13,
			Height: 102,
			Address: storage.Address{
				Id:         13,
				Height:     102,
				LastHeight: 102,
				Address:    "0x4444444444444444444444444444444444444444",
				IsContract: true,
			},
		},
		Implementation: nil,
	}
)

// ProxyContractHandlerTestSuite -
type ProxyContractHandlerTestSuite struct {
	suite.Suite
	contracts *mock.MockIProxyContract
	addresses *mock.MockIAddress
	echo      *echo.Echo
	handler   *ProxyContractHandler
	ctrl      *gomock.Controller
}

// SetupSuite -
func (s *ProxyContractHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.contracts = mock.NewMockIProxyContract(s.ctrl)
	s.addresses = mock.NewMockIAddress(s.ctrl)
	s.handler = NewProxyContractHandler(s.contracts, s.addresses, testIndexerName)
}

// TearDownSuite -
func (s *ProxyContractHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteProxyContractHandler_Run(t *testing.T) {
	suite.Run(t, new(ProxyContractHandlerTestSuite))
}

// TestListSuccess tests successful retrieval of proxy contracts list
func (s *ProxyContractHandlerTestSuite) TestListSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		Return([]storage.ProxyContract{
			testProxyContract1,
			testProxyContract2,
			testProxyContract3,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 3)

	// Check first contract
	s.Require().EqualValues(100, contracts[0].Height)
	s.Require().Equal("0x1111111111111111111111111111111111111111", contracts[0].Contract)
	s.Require().Equal("EIP1967", contracts[0].Type)
	s.Require().Equal("resolved", contracts[0].Status)
	s.Require().NotNil(contracts[0].Implementation)
	s.Require().Equal("0x2222222222222222222222222222222222222222", *contracts[0].Implementation)

	// Check second contract (no implementation)
	s.Require().EqualValues(101, contracts[1].Height)
	s.Require().Equal("0x3333333333333333333333333333333333333333", contracts[1].Contract)
	s.Require().Equal("EIP1967", contracts[1].Type)
	s.Require().Equal("new", contracts[1].Status)
	s.Require().Nil(contracts[1].Implementation)

	// Check third contract
	s.Require().EqualValues(102, contracts[2].Height)
	s.Require().Equal("0x4444444444444444444444444444444444444444", contracts[2].Contract)
	s.Require().Equal("EIP7702", contracts[2].Type)
	s.Require().Equal("error", contracts[2].Status)
	s.Require().Nil(contracts[2].Implementation)
}

// TestListWithLimit tests list with custom limit parameter
func (s *ProxyContractHandlerTestSuite) TestListWithLimit() {
	q := make(url.Values)
	q.Set("limit", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filters storage.ListProxyFilters) ([]storage.ProxyContract, error) {
			s.Require().Equal(5, filters.Limit)
			return []storage.ProxyContract{testProxyContract1}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 1)
}

// TestListWithOffset tests list with offset parameter
func (s *ProxyContractHandlerTestSuite) TestListWithOffset() {
	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "5")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filters storage.ListProxyFilters) ([]storage.ProxyContract, error) {
			s.Require().Equal(10, filters.Limit)
			s.Require().Equal(5, filters.Offset)
			return []storage.ProxyContract{testProxyContract2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 1)
}

// TestListWithTypeFilter tests list with type filter
func (s *ProxyContractHandlerTestSuite) TestListWithTypeFilter() {
	q := make(url.Values)
	q.Set("type", "EIP1967,EIP1167")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filters storage.ListProxyFilters) ([]storage.ProxyContract, error) {
			s.Require().Len(filters.Type, 2)
			s.Require().Equal(types.EIP1967, filters.Type[0])
			s.Require().Equal(types.EIP1167, filters.Type[1])
			return []storage.ProxyContract{testProxyContract1, testProxyContract2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 2)
}

// TestListWithStatusFilter tests list with status filter
func (s *ProxyContractHandlerTestSuite) TestListWithStatusFilter() {
	q := make(url.Values)
	q.Set("status", "resolved,new")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filters storage.ListProxyFilters) ([]storage.ProxyContract, error) {
			s.Require().Len(filters.Status, 2)
			s.Require().Equal(types.Resolved, filters.Status[0])
			s.Require().Equal(types.New, filters.Status[1])
			return []storage.ProxyContract{testProxyContract1, testProxyContract2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 2)
}

// TestListWithHeightFilter tests list with height filter
func (s *ProxyContractHandlerTestSuite) TestListWithHeightFilter() {
	q := make(url.Values)
	q.Set("height", "100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filters storage.ListProxyFilters) ([]storage.ProxyContract, error) {
			s.Require().Equal(pkgTypes.Level(100), filters.Height)
			return []storage.ProxyContract{testProxyContract1}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 1)
	s.Require().EqualValues(100, contracts[0].Height)
}

// TestListWithImplementationFilter tests list with implementation address filter
func (s *ProxyContractHandlerTestSuite) TestListWithImplementationFilter() {
	q := make(url.Values)
	q.Set("implementation", "0x2222222222222222222222222222222222222222")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.addresses.EXPECT().
		ByHash(gomock.Any(), pkgTypes.MustDecodeHex("2222222222222222222222222222222222222222")).
		Return(testImplementationAddress, nil).
		Times(1)

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filters storage.ListProxyFilters) ([]storage.ProxyContract, error) {
			s.Require().Equal(uint64(11), filters.ImplementationId)
			return []storage.ProxyContract{testProxyContract1}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 1)
	s.Require().NotNil(contracts[0].Implementation)
}

// TestListAscOrder tests list with ascending sort order
func (s *ProxyContractHandlerTestSuite) TestListAscOrder() {
	q := make(url.Values)
	q.Set("sort", "asc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		Return([]storage.ProxyContract{
			testProxyContract1,
			testProxyContract2,
			testProxyContract3,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 3)
}

// TestListEmptyResult tests list when no contracts are found
func (s *ProxyContractHandlerTestSuite) TestListEmptyResult() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.contracts.EXPECT().
		FilteredList(gomock.Any(), gomock.Any()).
		Return([]storage.ProxyContract{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var contracts []responses.ProxyContract
	err := json.NewDecoder(rec.Body).Decode(&contracts)
	s.Require().NoError(err)
	s.Require().Len(contracts, 0)
}

// TestListInvalidType tests list with invalid proxy type
func (s *ProxyContractHandlerTestSuite) TestListInvalidType() {
	q := make(url.Values)
	q.Set("type", "invalid_type")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidStatus tests list with invalid proxy status
func (s *ProxyContractHandlerTestSuite) TestListInvalidStatus() {
	q := make(url.Values)
	q.Set("status", "invalid_status")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidImplementationAddress tests list with invalid implementation address
func (s *ProxyContractHandlerTestSuite) TestListInvalidImplementationAddress() {
	q := make(url.Values)
	q.Set("implementation", "invalid_address")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListMaxLimit tests list with limit exceeding maximum
func (s *ProxyContractHandlerTestSuite) TestListMaxLimit() {
	q := make(url.Values)
	q.Set("limit", "101")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListInvalidSort tests list with invalid sort parameter
func (s *ProxyContractHandlerTestSuite) TestListInvalidSort() {
	q := make(url.Values)
	q.Set("sort", "invalid")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

// TestListNegativeOffset tests list with negative offset
func (s *ProxyContractHandlerTestSuite) TestListNegativeOffset() {
	q := make(url.Values)
	q.Set("offset", "-1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/proxy-contracts")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Require().NotEmpty(e.Message)
}

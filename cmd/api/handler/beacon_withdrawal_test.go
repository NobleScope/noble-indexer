package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/cmd/api/helpers"
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/mock"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var (
	testBeaconWithdrawal1 = storage.BeaconWithdrawal{
		Id:             1,
		Height:         100,
		Time:           testTime,
		Index:          0,
		ValidatorIndex: 42,
		AddressId:      1,
		Amount:         decimal.NewFromInt(32000000000),
		Address: storage.Address{
			Id:   1,
			Hash: testAddressHex1,
		},
	}

	testBeaconWithdrawal2 = storage.BeaconWithdrawal{
		Id:             2,
		Height:         101,
		Time:           testTime,
		Index:          1,
		ValidatorIndex: 43,
		AddressId:      2,
		Amount:         decimal.NewFromInt(16000000000),
		Address: storage.Address{
			Id:   2,
			Hash: testAddressHex2,
		},
	}
)

// BeaconWithdrawalHandlerTestSuite -
type BeaconWithdrawalHandlerTestSuite struct {
	suite.Suite
	beaconWithdrawals *mock.MockIBeaconWithdrawal
	address           *mock.MockIAddress
	echo              *echo.Echo
	handler           *BeaconWithdrawalHandler
	ctrl              *gomock.Controller
}

// SetupSuite -
func (s *BeaconWithdrawalHandlerTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.beaconWithdrawals = mock.NewMockIBeaconWithdrawal(s.ctrl)
	s.address = mock.NewMockIAddress(s.ctrl)
	s.handler = NewBeaconWithdrawalHandler(s.beaconWithdrawals, s.address)
}

// TearDownSuite -
func (s *BeaconWithdrawalHandlerTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteBeaconWithdrawalHandler_Run(t *testing.T) {
	suite.Run(t, new(BeaconWithdrawalHandlerTestSuite))
}

// TestListSuccess tests successful retrieval of beacon withdrawals
func (s *BeaconWithdrawalHandlerTestSuite) TestListSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/beacon_withdrawals")

	s.beaconWithdrawals.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		Return([]*storage.BeaconWithdrawal{&testBeaconWithdrawal1, &testBeaconWithdrawal2}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.BeaconWithdrawal `json:"result"`
		Cursor string                       `json:"cursor"`
	}
	err := json.NewDecoder(rec.Body).Decode(&body)
	s.Require().NoError(err)
	s.Require().Len(body.Result, 2)
	s.Require().NotEmpty(body.Cursor)
}

// TestListWithCursor tests cursor-based pagination for beacon withdrawals
func (s *BeaconWithdrawalHandlerTestSuite) TestListWithCursor() {
	q := make(url.Values)
	q.Set("cursor", helpers.EncodeTimeIDCursor(testTime, 1))

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/beacon_withdrawals")

	s.beaconWithdrawals.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filter storage.BeaconWithdrawalListFilter) ([]*storage.BeaconWithdrawal, error) {
			s.Require().EqualValues(1, filter.CursorID)
			s.Require().False(filter.CursorTime.IsZero())
			return []*storage.BeaconWithdrawal{&testBeaconWithdrawal2}, nil
		}).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.BeaconWithdrawal `json:"result"`
		Cursor string                       `json:"cursor"`
	}
	err := json.NewDecoder(rec.Body).Decode(&body)
	s.Require().NoError(err)
	s.Require().Len(body.Result, 1)
	s.Require().NotEmpty(body.Cursor)

	cursorTime, cursorID, err := helpers.DecodeTimeIDCursor(body.Cursor)
	s.Require().NoError(err)
	s.Require().EqualValues(testBeaconWithdrawal2.Id, cursorID)
	s.Require().Equal(testBeaconWithdrawal2.Time.UTC(), cursorTime.UTC())
}

// TestListInvalidCursor tests handling of invalid cursor
func (s *BeaconWithdrawalHandlerTestSuite) TestListInvalidCursor() {
	q := make(url.Values)
	q.Set("cursor", "not-valid-base64!!!")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/beacon_withdrawals")

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// TestListEmptyCursorOnEmpty tests that empty result returns empty cursor
func (s *BeaconWithdrawalHandlerTestSuite) TestListEmptyCursorOnEmpty() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/beacon_withdrawals")

	s.beaconWithdrawals.EXPECT().
		Filter(gomock.Any(), gomock.Any()).
		Return([]*storage.BeaconWithdrawal{}, nil).
		Times(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var body struct {
		Result []responses.BeaconWithdrawal `json:"result"`
		Cursor string                       `json:"cursor"`
	}
	err := json.NewDecoder(rec.Body).Decode(&body)
	s.Require().NoError(err)
	s.Require().Empty(body.Cursor)
}

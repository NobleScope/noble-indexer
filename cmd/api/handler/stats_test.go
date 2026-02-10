package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/mock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// StatsTestSuite -
type StatsTestSuite struct {
	suite.Suite
	blockStats *mock.MockIBlockStats
	state      *mock.MockIState
	echo       *echo.Echo
	handler    *StatsHandler
	ctrl       *gomock.Controller
}

// SetupSuite -
func (s *StatsTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.blockStats = mock.NewMockIBlockStats(s.ctrl)
	s.state = mock.NewMockIState(s.ctrl)
	s.handler = NewStatsHandler(s.state, s.blockStats, "test-indexer")
}

// TearDownSuite -
func (s *StatsTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteStats_Run(t *testing.T) {
	suite.Run(t, new(StatsTestSuite))
}

func (s *StatsTestSuite) TestAvgBlockTime() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/stats/block_time")

	s.blockStats.EXPECT().
		AvgBlockTime(gomock.Any(), gomock.Any()).
		Return(123.456, nil).
		Times(1)

	s.state.EXPECT().
		ByName(gomock.Any(), "test-indexer").
		Return(storage.State{LastTime: time.Now().UTC()}, nil).
		Times(1)

	s.Require().NoError(s.handler.AvgBlockTime(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var blockTime float64
	err := json.NewDecoder(rec.Body).Decode(&blockTime)
	s.Require().NoError(err)
	s.Require().EqualValues(123.456, blockTime)
}

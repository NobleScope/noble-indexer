// SPDX-FileCopyrightText: 2025 Bb Strategy Pte. Ltd. <celenium@baking-bad.org>
// SPDX-License-Identifier: MIT

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/mock"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var (
	testIndexerName = "test_indexer"
	testTime        = time.Date(2023, 8, 1, 1, 1, 0, 0, time.UTC)
)

// StateTestSuite -
type StateTestSuite struct {
	suite.Suite
	state   *mock.MockIState
	echo    *echo.Echo
	handler *StateHandler
	ctrl    *gomock.Controller
}

// SetupSuite -
func (s *StateTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.state = mock.NewMockIState(s.ctrl)
	s.handler = NewStateHandler(s.state, testIndexerName)
}

// TearDownSuite -
func (s *StateTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteState_Run(t *testing.T) {
	suite.Run(t, new(StateTestSuite))
}

func (s *StateTestSuite) TestHead() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/head")

	blockHashStr := "0x41654df0e5dbbb944c1f9ee175eef54a6a3f97722f97d1b3a150bda0edbe9bca"
	blockHashBytes, err := types.HexFromString(blockHashStr)
	s.Require().NoError(err)

	s.state.EXPECT().
		ByName(gomock.Any(), testIndexerName).
		Return(storage.State{
			Id:            1,
			Name:          testIndexerName,
			LastHeight:    100,
			LastHash:      blockHashBytes,
			LastTime:      testTime,
			TotalTx:       1234,
			TotalAccounts: 123,
		}, nil).
		Times(1)

	s.Require().NoError(s.handler.Head(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var state responses.State
	err = json.NewDecoder(rec.Body).Decode(&state)
	s.Require().NoError(err)
	s.Require().EqualValues(1, state.Id)
	s.Require().EqualValues(testIndexerName, state.Name)
	s.Require().EqualValues(100, state.LastHeight)
	s.Require().EqualValues(blockHashStr, state.LastHash)
	s.Require().Equal(testTime, state.LastTime)
	s.Require().EqualValues(1234, state.TotalTx)
	s.Require().EqualValues(123, state.TotalAccounts)
}

func (s *StateTestSuite) TestNoHead() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/head")

	s.state.EXPECT().
		ByName(gomock.Any(), testIndexerName).
		Return(storage.State{}, sql.ErrNoRows).
		Times(1)

	s.state.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true).
		Times(1)

	s.Require().NoError(s.handler.Head(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

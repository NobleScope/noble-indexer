// SPDX-FileCopyrightText: 2025 Bb Strategy Pte. Ltd. <celenium@baking-bad.org>
// SPDX-License-Identifier: MIT

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
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var (
	testBlock          storage.Block
	testBlockWithStats storage.Block
)

func init() {
	testBlock = storage.Block{
		Id:                   1,
		Height:               100,
		Time:                 testTime,
		GasLimit:             decimal.NewFromInt(1000000),
		GasUsed:              decimal.NewFromInt(500000),
		Hash:                 pkgTypes.Hex{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F},
		ParentHashHash:       pkgTypes.Hex{0x07, 0x64, 0x01, 0x22, 0x70, 0xaf, 0xac, 0xd3, 0xb1, 0x01, 0xbc, 0xfa, 0xda, 0xaa, 0x9f, 0xc3, 0x19, 0x0d, 0x04, 0xed, 0x90, 0xff, 0x22, 0xc0, 0xee, 0x59, 0x78, 0x1e, 0x54, 0x85, 0x8a, 0x7d},
		DifficultyHash:       pkgTypes.Hex{0x00},
		ExtraDataHash:        pkgTypes.Hex{0x72, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2e, 0x37, 0x2e, 0x30, 0x2f, 0x6c, 0x69, 0x6e, 0x75, 0x78},
		LogsBloomHash:        pkgTypes.Hex{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00},
		MixHash:              pkgTypes.Hex{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x33, 0xa8, 0x7e},
		NonceHash:            pkgTypes.Hex{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		ReceiptsRootHash:     pkgTypes.Hex{0x24, 0xe9, 0xaa, 0xe3, 0x03, 0x3f, 0x9f, 0xf8, 0x09, 0x67, 0x58, 0x31, 0xec, 0xa3, 0x31, 0xb7, 0x01, 0x44, 0x00, 0x09, 0x59, 0x2a, 0x40, 0xa6, 0xd7, 0x88, 0x75, 0x6f, 0x3b, 0xe9, 0x83, 0xa2},
		Sha3UnclesHash:       pkgTypes.Hex{0x1d, 0xcc, 0x4d, 0xe8, 0xde, 0xc7, 0x5d, 0x7a, 0xab, 0x85, 0xb5, 0x67, 0xb6, 0xcc, 0xd4, 0x1a, 0xd3, 0x12, 0x45, 0x1b, 0x94, 0x8a, 0x74, 0x13, 0xf0, 0xa1, 0x42, 0xfd, 0x40, 0xd4, 0x93, 0x47},
		SizeHash:             pkgTypes.Hex{0x58, 0x80},
		StateRootHash:        pkgTypes.Hex{0x9b, 0x6e, 0x76, 0xe8, 0x26, 0x3c, 0x50, 0x60, 0xb6, 0x1e, 0x39, 0x6c, 0x65, 0xba, 0xf1, 0x5d, 0xd1, 0x87, 0x38, 0x6d, 0x56, 0x07, 0x25, 0x0b, 0xe0, 0xdc, 0xc5, 0x30, 0x8f, 0x0b, 0x49, 0xef},
		TotalDifficultyHash:  pkgTypes.Hex{0x00},
		TransactionsRootHash: pkgTypes.Hex{0x07, 0x64, 0x01, 0x22, 0x70, 0xaf, 0xac, 0xd3, 0xb1, 0x01, 0xbc, 0xfa, 0xda, 0xaa, 0x9f, 0xc3, 0x19, 0x0d, 0x04, 0xed, 0x90, 0xff, 0x22, 0xc0, 0xee, 0x59, 0x78, 0x1e, 0x54, 0x85, 0x8a, 0x7d},
	}

	testBlockWithStats = testBlock
	testBlockWithStats.Stats = &storage.BlockStats{
		Id:        1,
		Height:    100,
		Time:      testTime,
		TxCount:   5,
		BlockTime: 100,
	}
}

// BlockTestSuite -
type BlockTestSuite struct {
	suite.Suite
	block      *mock.MockIBlock
	blockStats *mock.MockIBlockStats
	txs        *mock.MockITx
	state      *mock.MockIState
	echo       *echo.Echo
	handler    *BlockHandler
	ctrl       *gomock.Controller
}

// SetupSuite -
func (s *BlockTestSuite) SetupSuite() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()
	s.ctrl = gomock.NewController(s.T())
	s.block = mock.NewMockIBlock(s.ctrl)
	s.blockStats = mock.NewMockIBlockStats(s.ctrl)
	s.txs = mock.NewMockITx(s.ctrl)
	s.state = mock.NewMockIState(s.ctrl)
	s.handler = NewBlockHandler(s.block, s.blockStats, s.txs, s.state, testIndexerName)
}

// TearDownSuite -
func (s *BlockTestSuite) TearDownSuite() {
	s.ctrl.Finish()
	s.Require().NoError(s.echo.Shutdown(context.Background()))
}

func TestSuiteBlock_Run(t *testing.T) {
	suite.Run(t, new(BlockTestSuite))
}

func (s *BlockTestSuite) TestGet() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.block.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), false).
		Return(testBlock, nil)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var block responses.Block
	err := json.NewDecoder(rec.Body).Decode(&block)
	s.Require().NoError(err)
	s.Require().EqualValues(100, block.Height)
	s.Require().Equal(testTime, block.Time)
	s.Require().Equal("1000000", block.GasLimit.String())
	s.Require().Equal("500000", block.GasUsed.String())
	s.Require().Equal("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", block.Hash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", block.ParentHash)
	s.Require().Equal("0x00", block.Difficulty)
	s.Require().Equal("0x726574682f76312e372e302f6c696e7578", block.ExtraData)
	s.Require().Equal("0x000000000000000000002000000000", block.LogsBloom)
	s.Require().Equal("0x000000000000000000000000000000000000000000000000000000000033a87e", block.MixHash)
	s.Require().Equal(uint64(0), block.Nonce)
	s.Require().Equal("0x24e9aae3033f9ff809675831eca331b701440009592a40a6d788756f3be983a2", block.ReceiptsRoot)
	s.Require().Equal("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", block.Sha3Uncles)
	s.Require().Equal(uint64(22656), block.Size)
	s.Require().Equal("0x9b6e76e8263c5060b61e396c65baf15dd187386d5607250be0dcc5308f0b49ef", block.StateRoot)
	s.Require().Equal("0x00", block.TotalDifficultyHash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", block.TransactionsRootHash)
	s.Require().Nil(block.Stats)
}

func (s *BlockTestSuite) TestGetWithStats() {
	req := httptest.NewRequest(http.MethodGet, "/?stats=true", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.block.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), true).
		Return(testBlockWithStats, nil)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var block responses.Block
	err := json.NewDecoder(rec.Body).Decode(&block)
	s.Require().NoError(err)
	s.Require().EqualValues(100, block.Height)
	s.Require().Equal(testTime, block.Time)
	s.Require().Equal("1000000", block.GasLimit.String())
	s.Require().Equal("500000", block.GasUsed.String())
	s.Require().Equal("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", block.Hash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", block.ParentHash)
	s.Require().Equal("0x00", block.Difficulty)
	s.Require().Equal("0x726574682f76312e372e302f6c696e7578", block.ExtraData)
	s.Require().Equal("0x000000000000000000002000000000", block.LogsBloom)
	s.Require().Equal("0x000000000000000000000000000000000000000000000000000000000033a87e", block.MixHash)
	s.Require().Equal(uint64(0), block.Nonce)
	s.Require().Equal("0x24e9aae3033f9ff809675831eca331b701440009592a40a6d788756f3be983a2", block.ReceiptsRoot)
	s.Require().Equal("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", block.Sha3Uncles)
	s.Require().Equal(uint64(22656), block.Size)
	s.Require().Equal("0x9b6e76e8263c5060b61e396c65baf15dd187386d5607250be0dcc5308f0b49ef", block.StateRoot)
	s.Require().Equal("0x00", block.TotalDifficultyHash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", block.TransactionsRootHash)
	s.Require().NotNil(block.Stats)
	s.Require().EqualValues(100, block.Stats.Height)
	s.Require().EqualValues(testTime, block.Stats.Time)
	s.Require().EqualValues(5, block.Stats.TxCount)
	s.Require().EqualValues(100, block.Stats.BlockTime)
}

func (s *BlockTestSuite) TestGetNoContent() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height")
	c.SetParamNames("height")
	c.SetParamValues("100")

	s.block.EXPECT().
		ByHeight(gomock.Any(), pkgTypes.Level(100), false).
		Return(storage.Block{}, sql.ErrNoRows)

	s.block.EXPECT().
		IsNoRows(gomock.Any()).
		Return(true)

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusNoContent, rec.Code)
}

func (s *BlockTestSuite) TestGetInvalidBlockHeight() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block/:height")
	c.SetParamNames("height")
	c.SetParamValues("invalid")

	s.Require().NoError(s.handler.Get(c))
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var e Error
	err := json.NewDecoder(rec.Body).Decode(&e)
	s.Require().NoError(err)
	s.Contains(e.Message, "parsing")
}

func (s *BlockTestSuite) TestList() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block")

	s.block.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*storage.Block{
			&testBlock,
		}, nil).
		MaxTimes(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var blocks []responses.Block
	err := json.NewDecoder(rec.Body).Decode(&blocks)
	s.Require().NoError(err)
	s.Require().Len(blocks, 1)
	s.Require().EqualValues(100, blocks[0].Height)
	s.Require().Equal(testTime, blocks[0].Time)
	s.Require().Equal("1000000", blocks[0].GasLimit.String())
	s.Require().Equal("500000", blocks[0].GasUsed.String())
	s.Require().Equal("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", blocks[0].Hash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", blocks[0].ParentHash)
	s.Require().Equal("0x00", blocks[0].Difficulty)
	s.Require().Equal("0x726574682f76312e372e302f6c696e7578", blocks[0].ExtraData)
	s.Require().Equal("0x000000000000000000002000000000", blocks[0].LogsBloom)
	s.Require().Equal("0x000000000000000000000000000000000000000000000000000000000033a87e", blocks[0].MixHash)
	s.Require().Equal(uint64(0), blocks[0].Nonce)
	s.Require().Equal("0x24e9aae3033f9ff809675831eca331b701440009592a40a6d788756f3be983a2", blocks[0].ReceiptsRoot)
	s.Require().Equal("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", blocks[0].Sha3Uncles)
	s.Require().Equal(uint64(22656), blocks[0].Size)
	s.Require().Equal("0x9b6e76e8263c5060b61e396c65baf15dd187386d5607250be0dcc5308f0b49ef", blocks[0].StateRoot)
	s.Require().Equal("0x00", blocks[0].TotalDifficultyHash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", blocks[0].TransactionsRootHash)
}

func (s *BlockTestSuite) TestListWithStats() {
	q := make(url.Values)
	q.Set("stats", "true")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetPath("/block")

	s.block.EXPECT().
		ListWithStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*storage.Block{
			&testBlockWithStats,
		}, nil).
		MaxTimes(1)

	s.Require().NoError(s.handler.List(c))
	s.Require().Equal(http.StatusOK, rec.Code)

	var blocks []responses.Block
	err := json.NewDecoder(rec.Body).Decode(&blocks)
	s.Require().NoError(err)
	s.Require().Len(blocks, 1)
	s.Require().EqualValues(100, blocks[0].Height)
	s.Require().Equal(testTime, blocks[0].Time)
	s.Require().Equal("1000000", blocks[0].GasLimit.String())
	s.Require().Equal("500000", blocks[0].GasUsed.String())
	s.Require().Equal("0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", blocks[0].Hash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", blocks[0].ParentHash)
	s.Require().Equal("0x00", blocks[0].Difficulty)
	s.Require().Equal("0x726574682f76312e372e302f6c696e7578", blocks[0].ExtraData)
	s.Require().Equal("0x000000000000000000002000000000", blocks[0].LogsBloom)
	s.Require().Equal("0x000000000000000000000000000000000000000000000000000000000033a87e", blocks[0].MixHash)
	s.Require().Equal(uint64(0), blocks[0].Nonce)
	s.Require().Equal("0x24e9aae3033f9ff809675831eca331b701440009592a40a6d788756f3be983a2", blocks[0].ReceiptsRoot)
	s.Require().Equal("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", blocks[0].Sha3Uncles)
	s.Require().Equal(uint64(22656), blocks[0].Size)
	s.Require().Equal("0x9b6e76e8263c5060b61e396c65baf15dd187386d5607250be0dcc5308f0b49ef", blocks[0].StateRoot)
	s.Require().Equal("0x00", blocks[0].TotalDifficultyHash)
	s.Require().Equal("0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d", blocks[0].TransactionsRootHash)
	s.Require().NotNil(blocks[0].Stats)
	s.Require().EqualValues(100, blocks[0].Stats.Height)
	s.Require().EqualValues(testTime, blocks[0].Stats.Time)
	s.Require().EqualValues(5, blocks[0].Stats.TxCount)
	s.Require().EqualValues(100, blocks[0].Stats.BlockTime)
}

package handler

import (
	"net/http"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
)

type BlockHandler struct {
	block       storage.IBlock
	blockStats  storage.IBlockStats
	state       storage.IState
	indexerName string
}

func NewBlockHandler(
	block storage.IBlock,
	blockStats storage.IBlockStats,
	state storage.IState,
	indexerName string,
) *BlockHandler {
	return &BlockHandler{
		block:       block,
		blockStats:  blockStats,
		state:       state,
		indexerName: indexerName,
	}
}

type getBlockByHeightRequest struct {
	Height types.Level `param:"height" validate:"min=1"`
	Stats  bool        `query:"stats"  validate:"omitempty"`
}

// Get godoc
//
//	@Summary		Get block info
//	@Description	Get block info
//	@Tags			block
//	@ID				get-block
//	@Param			height	path	integer	true	"Block height"	minimum(1)
//	@Param			stats	query	boolean	false	"Need include stats for block"
//	@Produce		JSON
//	@Success		200	{object}	responses.Block
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/block/{height} [get]
func (handler *BlockHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getBlockByHeightRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	block, err := handler.block.ByHeight(c.Request().Context(), req.Height, req.Stats)
	if err != nil {
		return handleError(c, err, handler.block)
	}

	return c.JSON(http.StatusOK, responses.NewBlock(block))
}

type blockListRequest struct {
	Limit  uint64 `query:"limit"  validate:"omitempty,min=1,max=100"`
	Offset uint64 `query:"offset" validate:"omitempty,min=0"`
	Sort   string `query:"sort"   validate:"omitempty,oneof=asc desc"`
	Stats  bool   `query:"stats"  validate:"omitempty"`
}

func (p *blockListRequest) SetDefault() {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Sort == "" {
		p.Sort = asc
	}
}

// List godoc
//
//	@Summary		List blocks info
//	@Description	List blocks info
//	@Tags			block
//	@ID				list-block
//	@Param			limit	query	integer	false	"Count of requested blocks" 	minimum(1)	maximum(100)
//	@Param			offset	query	integer	false	"Offset"						minimum(1)
//	@Param			sort	query	string	false	"Sort order"					Enums(asc, desc)
//	@Param			stats	query	boolean	false	"Need join stats for block"
//	@Produce		JSON
//	@Success		200	{array}		responses.Block
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/block [get]
func (handler *BlockHandler) List(c echo.Context) error {
	req, err := bindAndValidate[blockListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	var blocks []*storage.Block
	if req.Stats {
		blocks, err = handler.block.ListWithStats(c.Request().Context(), req.Limit, req.Offset, pgSort(req.Sort))
	} else {
		blocks, err = handler.block.List(c.Request().Context(), req.Limit, req.Offset, pgSort(req.Sort))
	}

	if err != nil {
		return handleError(c, err, handler.block)
	}

	response := make([]responses.Block, len(blocks))
	for i := range blocks {
		response[i] = responses.NewBlock(*blocks[i])
	}

	return returnArray(c, response)
}

// Count godoc
//
//	@Summary		Get count of blocks in network
//	@Description	Get count of blocks in network
//	@Tags			block
//	@ID				get-blocks-count
//	@Produce		JSON
//	@Success		200	{integer}	uint64
//	@Failure		500	{object}	Error
//	@Router			/block/count [get]
func (handler *BlockHandler) Count(c echo.Context) error {
	state, err := handler.state.ByName(c.Request().Context(), handler.indexerName)
	if err != nil {
		return handleError(c, err, handler.block)
	}
	return c.JSON(http.StatusOK, state.LastHeight+1) // + genesis block
}

// GetStats godoc
//
//	@Summary		Get block stats by height
//	@Description	Get block stats by height
//	@Tags			block
//	@ID				get-block-stats
//	@Param			height	path	integer	true	"Block height"	minimum(1)
//	@Produce		JSON
//	@Success		200	{object}	responses.BlockStats
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/block/{height}/stats [get]
func (handler *BlockHandler) GetStats(c echo.Context) error {
	req, err := bindAndValidate[getBlockByHeightRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	stats, err := handler.blockStats.ByHeight(c.Request().Context(), req.Height)
	if err != nil {
		return handleError(c, err, handler.block)
	}
	return c.JSON(http.StatusOK, responses.NewBlockStats(stats))
}

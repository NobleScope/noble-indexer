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
	txs         storage.ITx
	state       storage.IState
	indexerName string
}

func NewBlockHandler(
	block storage.IBlock,
	blockStats storage.IBlockStats,
	txs storage.ITx,
	state storage.IState,
	indexerName string,
) *BlockHandler {
	return &BlockHandler{
		block:       block,
		blockStats:  blockStats,
		txs:         txs,
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
//	@Produce		json
//	@Success		200	{object}	responses.Block
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/blocks/{height} [get]
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
//	@Param			offset	query	integer	false	"Offset"						minimum(0)
//	@Param			sort	query	string	false	"Sort order"					Enums(asc, desc)
//	@Param			stats	query	boolean	false	"Need join stats for block"
//	@Produce		json
//	@Success		200	{array}		responses.Block
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/blocks [get]
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
//	@Produce		json
//	@Success		200	{integer}	uint64
//	@Failure		500	{object}	Error
//	@Router			/blocks/count [get]
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
//	@Produce		json
//	@Success		200	{object}	responses.BlockStats
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/blocks/{height}/stats [get]
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

type transactionsListRequest struct {
	Height types.Level `param:"height" validate:"min=1"`
	Limit  int         `query:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int         `query:"offset" validate:"omitempty,min=0"`
	Sort   string      `query:"sort"   validate:"omitempty,oneof=asc desc"`
}

func (p *transactionsListRequest) SetDefault() {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Sort == "" {
		p.Sort = asc
	}
}

// TransactionsList godoc
//
//	@Summary		List block transactions
//	@Description	List block transactions
//	@Tags			block
//	@ID				list-block-transactions
//	@Param			height	path	integer	true	"Block height"	minimum(1)
//	@Param			limit	query	integer	false	"Count of requested transactions" 	minimum(1)	maximum(100)
//	@Param			offset	query	integer	false	"Offset"							minimum(0)
//	@Param			sort	query	string	false	"Sort order"						Enums(asc, desc)
//	@Produce		json
//	@Success		200	{array}		responses.Transaction
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/blocks/{height}/transactions [get]
func (handler *BlockHandler) TransactionsList(c echo.Context) error {
	req, err := bindAndValidate[transactionsListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	var txs []*storage.Tx
	txs, err = handler.txs.ByHeight(c.Request().Context(), req.Height, req.Limit, req.Offset, pgSort(req.Sort))

	if err != nil {
		return handleError(c, err, handler.block)
	}

	response := make([]responses.Transaction, len(txs))
	for i := range txs {
		response[i] = responses.NewTransaction(*txs[i])
	}

	return returnArray(c, response)
}

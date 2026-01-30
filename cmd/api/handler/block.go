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

type getBlockByHeightWithStatsRequest struct {
	Height types.Level `param:"height" validate:"min=1"`
	Stats  bool        `query:"stats"  validate:"omitempty"`
}

// Get godoc
//
//	@Summary		Get block by height
//	@Description	Returns detailed information about a specific block at the given height. Optionally includes statistical data such as transaction count, gas usage, and fees.
//	@Tags			block
//	@ID				get-block
//	@Param			height	path	integer	true	"Block height (block number)"														minimum(1)	example(12345)
//	@Param			stats	query	boolean	false	"Include block statistics (transaction count, gas usage, fees) (default: false)"	default(false)
//	@Produce		json
//	@Success		200	{object}	responses.Block	"Block information"
//	@Success		204								"Block not found"
//	@Failure		400	{object}	Error			"Invalid request parameters"
//	@Failure		500	{object}	Error			"Internal server error"
//	@Router			/blocks/{height} [get]
func (handler *BlockHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getBlockByHeightWithStatsRequest](c)
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
	Limit  int    `query:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int    `query:"offset" validate:"omitempty,min=0"`
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
//	@Summary		List blocks
//	@Description	Returns a paginated list of blocks. Blocks can be sorted by height in ascending or descending order. Optionally includes statistics for each block.
//	@Tags			block
//	@ID				list-block
//	@Param			limit	query	integer	false	"Number of blocks to return (default: 10)" 				minimum(1)	maximum(100)	default(10)
//	@Param			offset	query	integer	false	"Number of blocks to skip (default: 0)"					minimum(0)	default(0)
//	@Param			sort	query	string	false	"Sort order by height (default: asc)"					Enums(asc, desc)	default(asc)
//	@Param			stats	query	boolean	false	"Include statistics for each block (default: false)"	default(false)
//	@Produce		json
//	@Success		200	{array}		responses.Block	"List of blocks"
//	@Failure		400	{object}	Error			"Invalid request parameters"
//	@Failure		500	{object}	Error			"Internal server error"
//	@Router			/blocks [get]
func (handler *BlockHandler) List(c echo.Context) error {
	req, err := bindAndValidate[blockListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filters := storage.BlockListFilter{
		Limit:     req.Limit,
		Offset:    req.Offset,
		Sort:      pgSort(req.Sort),
		WithStats: req.Stats,
	}

	var blocks []storage.Block
	blocks, err = handler.block.Filter(c.Request().Context(), filters)

	if err != nil {
		return handleError(c, err, handler.block)
	}

	response := make([]responses.Block, len(blocks))
	for i := range blocks {
		response[i] = responses.NewBlock(blocks[i])
	}

	return returnArray(c, response)
}

// Count godoc
//
//	@Summary		Get total block count
//	@Description	Returns the total number of blocks indexed in the blockchain, including the genesis block
//	@Tags			block
//	@ID				get-blocks-count
//	@Produce		json
//	@Success		200	{integer}	uint64	"Total number of blocks (including genesis)"
//	@Failure		500	{object}	Error	"Internal server error"
//	@Router			/blocks/count [get]
func (handler *BlockHandler) Count(c echo.Context) error {
	state, err := handler.state.ByName(c.Request().Context(), handler.indexerName)
	if err != nil {
		return handleError(c, err, handler.block)
	}
	return c.JSON(http.StatusOK, state.LastHeight+1) // + genesis block
}

type getBlockByHeightRequest struct {
	Height types.Level `param:"height" validate:"min=1"`
}

// GetStats godoc
//
//	@Summary		Get block statistics
//	@Description	Returns statistical data for a specific block including transaction count, gas used, gas limit, base fee, and total fees
//	@Tags			block
//	@ID				get-block-stats
//	@Param			height	path	integer	true	"Block height (block number)"	minimum(1)	example(12345)
//	@Produce		json
//	@Success		200	{object}	responses.BlockStats	"Block statistics"
//	@Failure		400	{object}	Error					"Invalid block height"
//	@Failure		500	{object}	Error					"Internal server error"
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
//	@Summary		List transactions in block
//	@Description	Returns a paginated list of all transactions included in the specified block
//	@Tags			block
//	@ID				list-block-transactions
//	@Param			height	path	integer	true	"Block height (block number)" 						minimum(1)	example(12345)
//	@Param			limit	query	integer	false	"Number of transactions to return (default: 10)" 	minimum(1)	maximum(100)	default(10)
//	@Param			offset	query	integer	false	"Number of transactions to skip (default: 0)"		minimum(0)	default(0)
//	@Param			sort	query	string	false	"Sort order by index (default: asc)"				Enums(asc, desc)	default(asc)
//	@Produce		json
//	@Success		200	{array}		responses.Transaction	"List of transactions in the block"
//	@Failure		400	{object}	Error					"Invalid block height"
//	@Failure		500	{object}	Error					"Internal server error"
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

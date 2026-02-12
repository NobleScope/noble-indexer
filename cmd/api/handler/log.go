package handler

import (
	"time"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
)

type LogHandler struct {
	log     storage.ILog
	tx      storage.ITx
	address storage.IAddress
}

func NewLogHandler(
	log storage.ILog,
	tx storage.ITx,
	address storage.IAddress,
) *LogHandler {
	return &LogHandler{
		log:     log,
		tx:      tx,
		address: address,
	}
}

type logListRequest struct {
	Limit   int     `query:"limit"   validate:"omitempty,min=1,max=100"`
	Offset  int     `query:"offset"  validate:"omitempty,min=0"`
	Sort    string  `query:"sort"    validate:"omitempty,oneof=asc desc"`
	Height  *uint64 `query:"height"  validate:"omitempty,min=0"`
	TxHash  string  `query:"tx_hash" validate:"omitempty,tx_hash"`
	Address string  `query:"address" validate:"omitempty,address"`
	Decode  bool    `query:"decode"  validate:"omitempty"`

	From int64 `example:"1692892095" query:"time_from" swaggertype:"integer" validate:"omitempty,min=1"`
	To   int64 `example:"1692892095" query:"time_to"   swaggertype:"integer" validate:"omitempty,min=1"`
}

func (req *logListRequest) SetDefault() {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Sort == "" {
		req.Sort = desc
	}
}

// List godoc
//
//	@Summary		List event logs
//	@Description	Returns a paginated list of event logs emitted by smart contracts. Can be filtered by transaction, address, block height, or time range.
//	@Tags			transactions
//	@ID				list-transaction-log
//	@Param			tx_hash			query	string	false	"Filter by transaction hash (hexadecimal with 0x prefix)"	minlength(66)	maxlength(66)	example(0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef)
//	@Param			limit			query	integer	false	"Number of logs to return (default: 10)"					minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of logs to skip (default: 0)"						minimum(0)	default(0)
//	@Param			address			query	string	false	"Filter by contract address that emitted the log"			minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			height			query	integer	false	"Filter by block height"									minimum(1)	example(12345)
//	@Param			time_from		query	integer	false	"Filter by timestamp from (Unix timestamp)"					minimum(1)	example(1692892095)
//	@Param			time_to			query	integer	false	"Filter by timestamp to (Unix timestamp)"					minimum(1)	example(1692892095)
//	@Param			sort			query	string	false	"Sort order by timestamp (default: desc)"					Enums(asc, desc)	default(desc)
//	@Param			decode			query	boolean	false	"Decode log data and topics using contract ABI"				default(false)
//	@Produce		json
//	@Success		200	{array}		responses.Log	"List of event logs"
//	@Failure		400	{object}	Error			"Invalid request parameters"
//	@Failure		500	{object}	Error			"Internal server error"
//	@Router			/logs [get]
func (handler *LogHandler) List(c echo.Context) error {
	req, err := bindAndValidate[logListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filters := storage.LogListFilter{
		Limit:   req.Limit,
		Offset:  req.Offset,
		Sort:    pgSort(req.Sort),
		WithABI: req.Decode,
	}

	if req.TxHash != "" {
		hash, err := types.HexFromString(req.TxHash)
		if err != nil {
			return badRequestError(c, err)
		}

		tx, err := handler.tx.ByHash(c.Request().Context(), hash)
		if err != nil {
			return handleError(c, err, handler.tx)
		}

		filters.TxId = &tx.Id
	}

	if req.Address != "" {
		hash, err := types.HexFromString(req.Address)
		if err != nil {
			return badRequestError(c, err)
		}

		address, err := handler.address.ByHash(c.Request().Context(), hash)
		if err != nil {
			return handleError(c, err, handler.address)
		}

		filters.AddressId = &address.Id
	}

	if req.Height != nil {
		filters.Height = req.Height
	}

	if req.From > 0 {
		filters.TimeFrom = time.Unix(req.From, 0).UTC()
	}
	if req.To > 0 {
		filters.TimeTo = time.Unix(req.To, 0).UTC()
	}

	logs, err := handler.log.Filter(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.log)
	}

	response := make([]responses.Log, len(logs))
	for i := range logs {
		response[i] = responses.NewLog(logs[i])
	}

	return returnArray(c, response)
}

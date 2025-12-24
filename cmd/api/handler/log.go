package handler

import (
	"time"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
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

	From int64 `example:"1692892095" query:"from" swaggertype:"integer" validate:"omitempty,min=1,max=16725214800"`
	To   int64 `example:"1692892095" query:"to"   swaggertype:"integer" validate:"omitempty,min=1,max=16725214800"`
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
//	@Summary		List logs
//	@Description	List logs
//	@Tags			transactions
//	@ID				list-transaction-log
//	@Param			tx_hash			query	string	false	"Transaction hash in hexadecimal with 0x prefix"	minlength(66)	maxlength(66)
//	@Param			limit			query	integer	false	"Count of requested entities"						minimum(1)	maximum(100)
//	@Param			offset			query	integer	false	"Offset"											minimum(0)
//	@Param			address			query	string	false	"Address whose invocation generated this log"		minlength(42)	maxlength(42)
//	@Param			height			query	integer	false	"Block height"										minimum(1)
//	@Param			sort			query	string	false	"Sort order. Default: desc"							Enums(asc, desc)
//	@Produce		json
//	@Success		200	{array}		responses.Trace
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/logs [get]
func (handler *LogHandler) List(c echo.Context) error {
	req, err := bindAndValidate[logListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filters := storage.LogListFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   pgSort(req.Sort),
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

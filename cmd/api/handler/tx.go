package handler

import (
	"net/http"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	internalTypes "github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
)

type TxHandler struct {
	tx          storage.ITx
	trace       storage.ITrace
	address     storage.IAddress
	indexerName string
}

func NewTxHandler(
	tx storage.ITx,
	trace storage.ITrace,
	address storage.IAddress,
	indexerName string,
) *TxHandler {
	return &TxHandler{
		tx:          tx,
		trace:       trace,
		address:     address,
		indexerName: indexerName,
	}
}

type getTxRequest struct {
	Hash string `param:"hash" validate:"required,tx_hash"`
}

// Get godoc
//
//	@Summary		Get transaction by hash
//	@Description	Get transaction by hash
//	@Tags			transactions
//	@ID				get-transaction
//	@Param			hash	path	string	true	"Transaction hash in hexadecimal with 0x prefix"	minlength(66)	maxlength(66)
//	@Produce		json
//	@Success		200	{object}	responses.Transaction
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/tx/{hash} [get]
func (handler *TxHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getTxRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	hash, err := types.HexFromString(req.Hash)
	if err != nil {
		return badRequestError(c, err)
	}

	tx, err := handler.tx.ByHash(c.Request().Context(), hash)
	if err != nil {
		return handleError(c, err, handler.tx)
	}

	return c.JSON(http.StatusOK, responses.NewTransaction(tx))
}

type getTxTraces struct {
	TxHash      string      `query:"tx_hash"      validate:"omitempty,tx_hash"`
	Limit       int         `query:"limit"        validate:"omitempty,min=1,max=100"`
	Offset      int         `query:"offset"       validate:"omitempty,min=0"`
	AddressFrom string      `query:"address_from" validate:"omitempty,address"`
	AddressTo   string      `query:"address_to"   validate:"omitempty,address"`
	Contract    string      `query:"contract"     validate:"omitempty,address"`
	Height      *uint64     `query:"height"       validate:"omitempty,min=0"`
	Type        StringArray `query:"type"         validate:"omitempty,dive,trace_type"`
	Sort        string      `query:"sort"         validate:"omitempty,oneof=asc desc"`
}

func (req *getTxTraces) SetDefault() {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Sort == "" {
		req.Sort = desc
	}
}

// Traces godoc
//
//	@Summary		List transaction traces
//	@Description	List transaction traces
//	@Tags			transactions
//	@ID				list-transaction-traces
//	@Param			tx_hash			query	string	false	"Transaction hash in hexadecimal with 0x prefix"	minlength(66)	maxlength(66)
//	@Param			limit			query	integer	false	"Count of requested entities"						minimum(1)		maximum(100)
//	@Param			offset			query	integer	false	"Offset"											minimum(0)
//	@Param			address_from	query	string	false	"Address which initiate trace"						minlength(42)	maxlength(42)
//	@Param			address_to		query	string	false	"Address which receiving trace result"				minlength(42)	maxlength(42)
//	@Param			contract		query	string	false	"Called contract"									minlength(42)	maxlength(42)
//	@Param			height			query	integer	false	"Block height"										minimum(1)
//	@Param			type			query	string	false	"Comma-separated list of trace types"				Enums(call, delegatecall, staticcall, create, create2, selfdestruct, reward, suicide)
//	@Param			sort			query	string	false	"Sort order. Default: desc"							Enums(asc, desc)
//	@Produce		json
//	@Success		200	{array}		responses.Trace
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/traces [get]
func (handler *TxHandler) Traces(c echo.Context) error {
	req, err := bindAndValidate[getTxTraces](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	traceTypes := make([]internalTypes.TraceType, len(req.Type))
	for i := range traceTypes {
		traceTypes[i] = internalTypes.TraceType(req.Type[i])
	}

	filters := storage.TraceListFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   pgSort(req.Sort),
		Type:   traceTypes,
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

	if req.AddressFrom != "" {
		hash, err := types.HexFromString(req.AddressFrom)
		if err != nil {
			return badRequestError(c, err)
		}

		address, err := handler.address.ByHash(c.Request().Context(), hash)
		if err != nil {
			return handleError(c, err, handler.address)
		}

		filters.AddressFromId = &address.Id
	}

	if req.AddressTo != "" {
		hash, err := types.HexFromString(req.AddressTo)
		if err != nil {
			return badRequestError(c, err)
		}

		address, err := handler.address.ByHash(c.Request().Context(), hash)
		if err != nil {
			return handleError(c, err, handler.address)
		}

		filters.AddressToId = &address.Id
	}

	if req.Contract != "" {
		hash, err := types.HexFromString(req.Contract)
		if err != nil {
			return badRequestError(c, err)
		}

		address, err := handler.address.ByHash(c.Request().Context(), hash)
		if err != nil {
			return handleError(c, err, handler.address)
		}

		filters.ContractId = &address.Id
	}

	if req.Height != nil {
		filters.Height = req.Height
	}

	traces, err := handler.trace.Filter(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.trace)
	}

	response := make([]responses.Trace, len(traces))
	for i := range traces {
		response[i] = responses.NewTrace(traces[i])
	}

	return returnArray(c, response)
}

type listTxs struct {
	Limit  int    `query:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int    `query:"offset" validate:"omitempty,min=0"`
	Sort   string `query:"sort"   validate:"omitempty,oneof=asc desc"`
}

func (req *listTxs) SetDefault() {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Sort == "" {
		req.Sort = desc
	}
}

// List godoc
//
//	@Summary		Transactions list
//	@Description	List all of indexed transactions
//	@Tags			transactions
//	@ID				list-transactions
//	@Param			limit	query	integer	false	"Count of requested entities"	minimum(1)	maximum(100)
//	@Param			offset	query	integer	false	"Offset"						minimum(0)
//	@Param			sort	query	string	false	"Sort order. Default: desc"		Enums(asc, desc)
//	@Produce		json
//	@Success		200	{array}		responses.Transaction
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/tx [get]
func (handler *TxHandler) List(c echo.Context) error {
	req, err := bindAndValidate[listTxs](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	txs, err := handler.tx.List(c.Request().Context(), uint64(req.Limit), uint64(req.Offset), pgSort(req.Sort))
	if err != nil {
		return handleError(c, err, handler.trace)
	}

	response := make([]responses.Transaction, len(txs))
	for i := range txs {
		response[i] = responses.NewTransaction(*txs[i])
	}

	return returnArray(c, response)
}

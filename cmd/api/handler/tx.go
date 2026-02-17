package handler

import (
	"net/http"
	"time"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/internal/storage"
	internalTypes "github.com/NobleScope/noble-indexer/internal/storage/types"
	"github.com/NobleScope/noble-indexer/pkg/types"
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
	Hash   string `param:"hash"   validate:"required,tx_hash"`
	Decode bool   `query:"decode" validate:"omitempty"`
}

// Get godoc
//
//	@Summary		Get transaction by hash
//	@Description	Returns detailed information about a specific transaction including status, gas used, value transferred, and associated traces
//	@Tags			transactions
//	@ID				get-transaction
//	@Param			hash	path	string	true	"Transaction hash in hexadecimal with 0x prefix"	minlength(66)	maxlength(66)	example(0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef)
//	@Param			decode	query	boolean	false	"Decode transaction input using contract ABI"		default(false)
//	@Produce		json
//	@Success		200	{object}	responses.Transaction	"Transaction information"
//	@Success		204										"Transaction not found"
//	@Failure		400	{object}	Error					"Invalid transaction hash format"
//	@Failure		500	{object}	Error					"Internal server error"
//	@Router			/txs/{hash} [get]
func (handler *TxHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getTxRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	hash, err := types.HexFromString(req.Hash)
	if err != nil {
		return badRequestError(c, err)
	}

	tx, err := handler.tx.ByHash(c.Request().Context(), hash, req.Decode)
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
	Address     string      `query:"address"      validate:"omitempty,address"`
	Contract    string      `query:"contract"     validate:"omitempty,address"`
	Height      *uint64     `query:"height"       validate:"omitempty,min=0"`
	Type        StringArray `query:"type"         validate:"omitempty,dive,trace_type"`
	Sort        string      `query:"sort"         validate:"omitempty,oneof=asc desc"`
	Decode      bool        `query:"decode"       validate:"omitempty"`
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
//	@Summary		List execution traces
//	@Description	Returns a paginated list of execution traces showing internal calls, contract creations, and other EVM operations. Traces provide detailed insight into transaction execution. Can be filtered by transaction, addresses, contract, block height, or trace type.
//	@Tags			transactions
//	@ID				list-transaction-traces
//	@Param			tx_hash			query	string	false	"Filter by transaction hash (hexadecimal with 0x prefix)"	minlength(66)	maxlength(66)	example(0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef)
//	@Param			limit			query	integer	false	"Number of traces to return (default: 10)"					minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of traces to skip (default: 0)"						minimum(0)	default(0)
//	@Param			address_from	query	string	false	"Filter by initiator address"								minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			address_to		query	string	false	"Filter by target address"									minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			address			query	string	false	"Filter by address (from or to)"							minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			height			query	integer	false	"Filter by block height"									minimum(1)	example(12345)
//	@Param			type			query	string	false	"Filter by trace type (comma-separated list)"				Enums(call, delegatecall, staticcall, create, create2, selfdestruct, reward, suicide)
//	@Param			sort			query	string	false	"Sort order (default: desc)"								Enums(asc, desc)	default(desc)
//	@Param			decode			query	boolean	false	"Decode trace input using contract ABI"						default(false)
//	@Produce		json
//	@Success		200	{array}		responses.Trace	"List of execution traces"
//	@Failure		400	{object}	Error			"Invalid request parameters"
//	@Failure		500	{object}	Error			"Internal server error"
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
		Limit:   req.Limit,
		Offset:  req.Offset,
		Sort:    pgSort(req.Sort),
		Type:    traceTypes,
		Height:  req.Height,
		WithABI: req.Decode,
	}

	if req.TxHash != "" {
		hash, err := types.HexFromString(req.TxHash)
		if err != nil {
			return badRequestError(c, err)
		}

		tx, err := handler.tx.ByHash(c.Request().Context(), hash, false)
		if err != nil {
			return handleError(c, err, handler.tx)
		}

		filters.TxId = &tx.Id
	}

	if req.AddressFrom != "" {
		address, err := handler.getAddressByHash(c, req.AddressFrom)
		if err != nil {
			return err
		}
		filters.AddressFromId = &address.Id
	}

	if req.AddressTo != "" {
		address, err := handler.getAddressByHash(c, req.AddressTo)
		if err != nil {
			return err
		}
		filters.AddressToId = &address.Id
	}

	if req.Address != "" {
		address, err := handler.getAddressByHash(c, req.Address)
		if err != nil {
			return err
		}
		filters.AddressId = &address.Id
	}

	if req.Contract != "" {
		address, err := handler.getAddressByHash(c, req.Contract)
		if err != nil {
			return err
		}
		filters.ContractId = &address.Id
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
	Limit       int         `query:"limit"        validate:"omitempty,min=1,max=100"`
	Offset      int         `query:"offset"       validate:"omitempty,min=0"`
	Sort        string      `query:"sort"         validate:"omitempty,oneof=asc desc"`
	AddressFrom string      `query:"address_from" validate:"omitempty,address"`
	AddressTo   string      `query:"address_to"   validate:"omitempty,address"`
	Address     string      `query:"address"      validate:"omitempty,address"`
	Contract    string      `query:"contract"     validate:"omitempty,address"`
	Height      *uint64     `query:"height"       validate:"omitempty,min=0"`
	Type        StringArray `query:"type"         validate:"omitempty,dive,tx_type"`
	Status      StringArray `query:"status"       validate:"omitempty,dive,tx_status"`
	Decode      bool        `query:"decode"       validate:"omitempty"`

	From int64 `example:"1692892095" query:"time_from" swaggertype:"integer" validate:"omitempty,min=1"`
	To   int64 `example:"1692892095" query:"time_to"   swaggertype:"integer" validate:"omitempty,min=1"`
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
//	@Summary		List transactions
//	@Description	Returns a paginated list of blockchain transactions. Can be filtered by addresses, contract, block height, transaction type, status, or time range. Supports various transaction types including legacy, EIP-1559 (dynamic fee), EIP-4844 (blob), and EIP-7702 (set code).
//	@Tags			transactions
//	@ID				list-transactions
//	@Param			limit			query	integer	false	"Number of transactions to return (default: 10)"	minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of transactions to skip (default: 0)"		minimum(0)	default(0)
//	@Param			address_from	query	string	false	"Filter by sender address"							minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			address_to		query	string	false	"Filter by recipient address"						minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			contract		query	string	false	"Filter by called contract address"					minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			height			query	integer	false	"Filter by block height"							minimum(1)	example(12345)
//	@Param			type			query	string	false	"Filter by transaction type (comma-separated list)"	Enums(TxTypeUnknown, TxTypeLegacy, TxTypeDynamicFee, TxTypeBlob, TxTypeSetCode)
//	@Param			status			query	string	false	"Filter by execution status (comma-separated list)"	Enums(TxStatusSuccess, TxStatusRevert)
//	@Param			time_from		query	integer	false	"Filter by timestamp from (Unix timestamp)"			minimum(1)	example(1692892095)
//	@Param			time_to			query	integer	false	"Filter by timestamp to (Unix timestamp)"			minimum(1)	example(1692892095)
//	@Param			sort			query	string	false	"Sort order by timestamp (default: desc)"			Enums(asc, desc)	default(desc)
//	@Param			decode			query	boolean	false	"Decode transaction input using contract ABI"		default(false)
//	@Produce		json
//	@Success		200	{array}		responses.Transaction	"List of transactions"
//	@Failure		400	{object}	Error					"Invalid request parameters"
//	@Failure		500	{object}	Error					"Internal server error"
//	@Router			/txs [get]
func (handler *TxHandler) List(c echo.Context) error {
	req, err := bindAndValidate[listTxs](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	txTypes := make([]internalTypes.TxType, len(req.Type))
	for i := range txTypes {
		txTypes[i] = internalTypes.TxType(req.Type[i])
	}

	txStatus := make([]internalTypes.TxStatus, len(req.Status))
	for i := range txStatus {
		txStatus[i] = internalTypes.TxStatus(req.Status[i])
	}

	filters := storage.TxListFilter{
		Limit:   req.Limit,
		Offset:  req.Offset,
		Sort:    pgSort(req.Sort),
		Type:    txTypes,
		Status:  txStatus,
		Height:  req.Height,
		WithABI: req.Decode,
	}

	if req.AddressFrom != "" {
		address, err := handler.getAddressByHash(c, req.AddressFrom)
		if err != nil {
			return err
		}
		filters.AddressFromId = &address.Id
	}

	if req.AddressTo != "" {
		address, err := handler.getAddressByHash(c, req.AddressTo)
		if err != nil {
			return err
		}
		filters.AddressToId = &address.Id
	}

	if req.Address != "" {
		address, err := handler.getAddressByHash(c, req.Address)
		if err != nil {
			return err
		}
		filters.AddressId = &address.Id
	}

	if req.Contract != "" {
		address, err := handler.getAddressByHash(c, req.Contract)
		if err != nil {
			return err
		}
		filters.ContractId = &address.Id
	}

	if req.From > 0 {
		filters.TimeFrom = time.Unix(req.From, 0).UTC()
	}
	if req.To > 0 {
		filters.TimeTo = time.Unix(req.To, 0).UTC()
	}

	txs, err := handler.tx.Filter(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.tx)
	}

	response := make([]responses.Transaction, len(txs))
	for i := range txs {
		response[i] = responses.NewTransaction(txs[i])
	}

	return returnArray(c, response)
}

func (handler *TxHandler) getAddressByHash(c echo.Context, h string) (storage.Address, error) {
	hash, err := types.HexFromString(h)
	if err != nil {
		return storage.Address{}, badRequestError(c, err)
	}

	address, err := handler.address.ByHash(c.Request().Context(), hash)
	if err != nil {
		return address, handleError(c, err, handler.address)
	}

	return address, nil
}

type getTxTracesTreeRequest struct {
	Hash   string `param:"hash"   validate:"required,tx_hash"`
	Decode bool   `query:"decode" validate:"omitempty"`
}

// TxTracesTree godoc
//
//	@Summary		Get transaction execution trace tree
//	@Description	Returns the execution trace tree for a specific transaction, showing all internal calls, contract creations, and EVM operations in a hierarchical structure. Each trace includes details such as type, gas used, value transferred, and any errors encountered during execution.
//	@Tags			transactions
//	@ID				get-transaction-trace-tree
//	@Param			hash	path	string	true	"Transaction hash in hexadecimal with 0x prefix"	minlength(66)	maxlength(66)	example(0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef)
//	@Param			decode	query	boolean	false	"Decode trace input using contract ABI"				default(false)
//	@Produce		json
//	@Success		200	{object}	responses.TraceTreeItem	"Execution trace tree for the transaction"
//	@Success		204										"Transaction not found or no traces available"
//	@Failure		400	{object}	Error					"Invalid transaction hash format"
//	@Failure		500	{object}	Error					"Internal server error"
//	@Router			/txs/{hash}/traces_tree [get]
func (handler *TxHandler) TxTracesTree(c echo.Context) error {
	req, err := bindAndValidate[getTxTracesTreeRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	hash, err := types.HexFromString(req.Hash)
	if err != nil {
		return badRequestError(c, err)
	}

	tx, err := handler.tx.ByHash(c.Request().Context(), hash, false)
	if err != nil {
		return handleError(c, err, handler.tx)
	}

	traces, err := handler.trace.ByTxId(c.Request().Context(), tx.Id, req.Decode)
	if err != nil {
		return handleError(c, err, handler.trace)
	}

	response, err := responses.BuildTraceTree(traces)
	if err != nil {
		return handleError(c, err, handler.trace)
	}

	return c.JSON(http.StatusOK, response)
}

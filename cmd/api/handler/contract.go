package handler

import (
	"net/http"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
)

type ContractHandler struct {
	contract storage.IContract
	tx       storage.ITx
	source   storage.ISource
}

func NewContractHandler(
	contract storage.IContract,
	tx storage.ITx,
	source storage.ISource,
) *ContractHandler {
	return &ContractHandler{
		contract: contract,
		tx:       tx,
		source:   source,
	}
}

type contractListRequest struct {
	Limit      int    `query:"limit"       validate:"omitempty,min=1,max=100"`
	Offset     int    `query:"offset"      validate:"omitempty,min=0"`
	Sort       string `query:"sort"        validate:"omitempty,oneof=asc desc"`
	SortBy     string `query:"sort_by"     validate:"omitempty,oneof=id height"`
	IsVerified bool   `query:"is_verified" validate:"omitempty"`
	TxHash     string `query:"tx_hash"     validate:"omitempty,tx_hash"`
}

func (p *contractListRequest) SetDefault() {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Sort == "" {
		p.Sort = asc
	}
}

// List godoc
//
//	@Summary		List smart contracts
//	@Description	Returns a paginated list of deployed smart contracts. Can be filtered by verification status or deployment transaction.
//	@Tags			contract
//	@ID				list-contract
//	@Param			limit		query	integer	false	"Number of contracts to return (default: 10)"								minimum(1)	maximum(100)	default(10)
//	@Param			offset		query	integer	false	"Number of contracts to skip (default: 0)"									minimum(0)	default(0)
//	@Param			sort		query	string	false	"Sort order (default: asc)"													Enums(asc, desc)	default(asc)
//	@Param			sort_by		query	string	false	"Field to sort by (default: id)"											Enums(id, height)
//	@Param			is_verified	query	boolean	false	"Filter to show only verified contracts (default: false)"					default(false)
//	@Param			tx_hash		query	string	false	"Filter by deployment transaction hash (hexadecimal with 0x prefix)"		minlength(66)	maxlength(66)	example(0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef)
//	@Produce		json
//	@Success		200	{array}		responses.Contract	"List of smart contracts"
//	@Failure		400	{object}	Error				"Invalid request parameters"
//	@Failure		500	{object}	Error				"Internal server error"
//	@Router			/contract [get]
func (handler *ContractHandler) List(c echo.Context) error {
	req, err := bindAndValidate[contractListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filters := storage.ContractListFilter{
		Limit:      req.Limit,
		Offset:     req.Offset,
		Sort:       pgSort(req.Sort),
		SortField:  req.SortBy,
		IsVerified: req.IsVerified,
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

	contracts, err := handler.contract.ListWithTx(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.contract)
	}

	response := make([]responses.Contract, len(contracts))
	for i := range contracts {
		response[i] = responses.NewContract(contracts[i])
	}

	return returnArray(c, response)
}

type getByHashRequest struct {
	Hash string `param:"hash" validate:"required,address"`
}

// Get godoc
//
//	@Summary		Get contract by address
//	@Description	Returns detailed information about a specific smart contract including deployment info, verification status, and metadata
//	@Tags			contract
//	@ID				get-contract
//	@Param			hash	path	string	true	"Contract address in hexadecimal format (e.g., 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)"	minlength(42)	maxlength(42)
//	@Produce		json
//	@Success		200	{object}	responses.Contract	"Contract information"
//	@Success		204						"Contract not found"
//	@Failure		400	{object}	Error			"Invalid contract address format"
//	@Failure		500	{object}	Error			"Internal server error"
//	@Router			/contract/{hash} [get]
func (handler *ContractHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getByHashRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	hash, err := types.HexFromString(req.Hash)
	if err != nil {
		return badRequestError(c, err)
	}

	contract, err := handler.contract.ByHash(c.Request().Context(), hash)
	if err != nil {
		return handleError(c, err, handler.contract)
	}

	return c.JSON(http.StatusOK, responses.NewContract(contract))
}

type getSourcesRequest struct {
	Hash   string `param:"hash"   validate:"required,address"`
	Limit  int    `query:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int    `query:"offset" validate:"omitempty,min=0"`
}

// ContractSources godoc
//
//	@Summary		Get contract source code
//	@Description	Returns the verified source code files for a specific smart contract. Only available for verified contracts.
//	@Tags			contract
//	@ID				get-contract-sources
//	@Param			hash	path	string	true	"Contract address in hexadecimal format (e.g., 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)"	minlength(42)	maxlength(42)
//	@Param			limit	query	integer	false	"Number of source files to return (default: 10)"											minimum(1)	maximum(100)	default(10)
//	@Param			offset	query	integer	false	"Number of source files to skip (default: 0)"												minimum(0)	default(0)
//	@Produce		json
//	@Success		200	{array}		responses.Source	"List of source code files"
//	@Success		204							"Contract not found or not verified"
//	@Failure		400	{object}	Error				"Invalid contract address format"
//	@Failure		500	{object}	Error				"Internal server error"
//	@Router			/contract/{hash}/sources [get]
func (handler *ContractHandler) ContractSources(c echo.Context) error {
	req, err := bindAndValidate[getSourcesRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	hash, err := types.HexFromString(req.Hash)
	if err != nil {
		return badRequestError(c, err)
	}

	contract, err := handler.contract.ByHash(c.Request().Context(), hash)
	if err != nil {
		return handleError(c, err, handler.contract)
	}

	sources, err := handler.source.ByContractId(c.Request().Context(), contract.Id, req.Limit, req.Offset)
	if err != nil {
		return handleError(c, err, handler.source)
	}

	response := make([]responses.Source, len(sources))
	for i := range sources {
		response[i] = responses.NewSource(sources[i])
	}

	return returnArray(c, response)
}

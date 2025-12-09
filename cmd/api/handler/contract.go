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
	TxHash     string `query:"tx_hash"     validate:"omitempty,txHash"`
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
//	@Summary		List contract info
//	@Description	List contract info
//	@Tags			contract
//	@ID				list-contract
//	@Param			limit	query	integer	false	"Count of requested entities"	mininum(1)	maximum(100)
//	@Param			offset	query	integer	false	"Offset"						mininum(1)
//	@Param			sort	query	string	false	"Sort order"					Enums(asc, desc)
//	@Param			sort_by	query	string	false	"Sort field"					Enums(id, height)
//	@Param			is_verified	query	boolean	false	"Show only verified contracts"
//	@Param			tx_hash	query	string	false	"Transaction hash in hexadecimal with 0x prefix"	minlength(66)	maxlength(66)
//	@Produce		json
//	@Success		200	{array}		responses.Contract
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
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

type getByTxHashRequest struct {
	Hash string `param:"hash" validate:"required,txHash"`
}

// Get godoc
//
//	@Summary		Get contract info
//	@Description	Get contract info
//	@Tags			contract
//	@ID				get-contract
//	@Param			hash	path	string	true	"Hash"	minlength(66)	maxlength(66)
//	@Produce		json
//	@Success		200	{object}	responses.Contract
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/contract/{hash} [get]
func (handler *ContractHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getByTxHashRequest](c)
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

	contract, err := handler.contract.ByTxId(c.Request().Context(), tx.Id)
	if err != nil {
		return handleError(c, err, handler.contract)
	}

	return c.JSON(http.StatusOK, responses.NewContract(contract))
}

type getSourcesRequest struct {
	Hash   string `param:"hash"   validate:"required,txHash"`
	Limit  int    `query:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int    `query:"offset" validate:"omitempty,min=0"`
}

// ContractSources godoc
//
//	@Summary		Get contract sources
//	@Description	Get contract sources
//	@Tags			contract
//	@ID				get-contract-sources
//	@Param			hash	path	string	true	"Hash"	minlength(66)	maxlength(66)
//	@Param			limit	query	integer	false	"Count of requested entities"	mininum(1)	maximum(100)
//	@Param			offset	query	integer	false	"Offset"						mininum(1)
//	@Produce		json
//	@Success		200	{object}	responses.Contract
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
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

	tx, err := handler.tx.ByHash(c.Request().Context(), hash)
	if err != nil {
		return handleError(c, err, handler.tx)
	}

	contract, err := handler.contract.ByTxId(c.Request().Context(), tx.Id)
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

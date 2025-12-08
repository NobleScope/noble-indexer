package handler

import (
	"net/http"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/labstack/echo/v4"
)

type AddressHandler struct {
	address     storage.IAddress
	txs         storage.ITx
	state       storage.IState
	indexerName string
}

func NewAddressHandler(
	address storage.IAddress,
	txs storage.ITx,
	state storage.IState,
	indexerName string,
) *AddressHandler {
	return &AddressHandler{
		address:     address,
		txs:         txs,
		state:       state,
		indexerName: indexerName,
	}
}

type addressListRequest struct {
	Limit         int    `query:"limit"          validate:"omitempty,min=1,max=100"`
	Offset        int    `query:"offset"         validate:"omitempty,min=0"`
	Sort          string `query:"sort"           validate:"omitempty,oneof=asc desc"`
	SortBy        string `query:"sort_by"        validate:"omitempty,oneof=id value first_height last_height"`
	OnlyContracts bool   `query:"only_contracts" validate:"omitempty"`
}

func (p *addressListRequest) SetDefault() {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Sort == "" {
		p.Sort = asc
	}
}

// List godoc
//
//	@Summary		List address info
//	@Description	List address info
//	@Tags			address
//	@ID				list-address
//	@Param			limit	query	integer	false	"Count of requested entities"	mininum(1)	maximum(100)
//	@Param			offset	query	integer	false	"Offset"						mininum(1)
//	@Param			sort	query	string	false	"Sort order"					Enums(asc, desc)
//	@Param			sort_by	query	string	false	"Sort field"					Enums(id, value, first_height, last_height)
//	@Param			only_contracts	query	boolean	false	"Show only contract addresses"
//	@Produce		json
//	@Success		200	{array}		responses.Address
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/address [get]
func (handler *AddressHandler) List(c echo.Context) error {
	req, err := bindAndValidate[addressListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	fltrs := storage.AddressListFilter{
		Limit:         req.Limit,
		Offset:        req.Offset,
		Sort:          pgSort(req.Sort),
		SortField:     req.SortBy,
		OnlyContracts: req.OnlyContracts,
	}

	address, err := handler.address.ListWithBalance(c.Request().Context(), fltrs)
	if err != nil {
		return handleError(c, err, handler.address)
	}

	response := make([]responses.Address, len(address))
	for i := range address {
		response[i] = responses.NewAddress(address[i])
	}

	return returnArray(c, response)
}

type getAddressRequest struct {
	Hash string `param:"hash" validate:"required,address"`
}

// Get godoc
//
//	@Summary		Get address info
//	@Description	Get address info
//	@Tags			address
//	@ID				get-address
//	@Param			hash	path	string	true	"Hash"	minlength(42)	maxlength(42)
//	@Produce		json
//	@Success		200	{object}	responses.Address
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/address/{hash} [get]
func (handler *AddressHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getAddressRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	address, err := handler.address.ByHash(c.Request().Context(), req.Hash)
	if err != nil {
		return handleError(c, err, handler.address)
	}

	return c.JSON(http.StatusOK, responses.NewAddress(address))
}

package handler

import (
	"net/http"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
)

type AddressHandler struct {
	address storage.IAddress
}

func NewAddressHandler(
	address storage.IAddress,
) *AddressHandler {
	return &AddressHandler{
		address: address,
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
//	@Summary		List addresses
//	@Description	Returns a paginated list of addresses with their balances. Can be filtered to show only contract addresses.
//	@Tags			address
//	@ID				list-address
//	@Param			limit			query	integer	false	"Number of addresses to return (default: 10)"								minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of addresses to skip (default: 0)"									minimum(0)	default(0)
//	@Param			sort			query	string	false	"Sort order (default: asc)"													Enums(asc, desc)	default(asc)
//	@Param			sort_by			query	string	false	"Field to sort by (default: id)"											Enums(id, value, last_height)
//	@Param			only_contracts	query	boolean	false	"If true, return only addresses that are smart contracts (default: false)"	default(false)
//	@Produce		json
//	@Success		200	{array}		responses.Address	"List of addresses with their balances"
//	@Failure		400	{object}	Error				"Invalid request parameters"
//	@Failure		500	{object}	Error				"Internal server error"
//	@Router			/addresses [get]
func (handler *AddressHandler) List(c echo.Context) error {
	req, err := bindAndValidate[addressListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filters := storage.AddressListFilter{
		Limit:         req.Limit,
		Offset:        req.Offset,
		Sort:          pgSort(req.Sort),
		SortField:     req.SortBy,
		OnlyContracts: req.OnlyContracts,
	}

	addresses, err := handler.address.ListWithBalance(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.address)
	}

	response := make([]responses.Address, len(addresses))
	for i := range addresses {
		response[i] = responses.NewAddress(addresses[i])
	}

	return returnArray(c, response)
}

type getAddressRequest struct {
	Hash string `param:"hash" validate:"required,address"`
}

// Get godoc
//
//	@Summary		Get address by hash
//	@Description	Returns detailed information about a specific address including its balance, contract status, and activity history
//	@Tags			address
//	@ID				get-address
//	@Param			hash	path	string	true	"Address hash in hexadecimal format (e.g., 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)"	minlength(42)	maxlength(42)
//	@Produce		json
//	@Success		200	{object}	responses.Address	"Address information with balance"
//	@Success		204									"Address not found"
//	@Failure		400	{object}	Error				"Invalid address hash format"
//	@Failure		500	{object}	Error				"Internal server error"
//	@Router			/addresses/{hash} [get]
func (handler *AddressHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[getAddressRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	hash, err := types.HexFromString(req.Hash)
	if err != nil {
		return badRequestError(c, err)
	}

	address, err := handler.address.ByHash(c.Request().Context(), hash)
	if err != nil {
		return handleError(c, err, handler.address)
	}

	return c.JSON(http.StatusOK, responses.NewAddress(address))
}

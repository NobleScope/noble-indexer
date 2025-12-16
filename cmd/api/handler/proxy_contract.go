package handler

import (
	"context"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type ProxyContractHandler struct {
	contracts   storage.IProxyContract
	addresses   storage.IAddress
	indexerName string
}

func NewProxyContractHandler(
	proxyContracts storage.IProxyContract,
	addresses storage.IAddress,
	indexerName string,
) *ProxyContractHandler {
	return &ProxyContractHandler{
		contracts:   proxyContracts,
		addresses:   addresses,
		indexerName: indexerName,
	}
}

type listProxyContracts struct {
	Limit          int         `query:"limit"          validate:"omitempty,min=1,max=100"`
	Offset         int         `query:"offset"         validate:"omitempty,min=0"`
	Sort           string      `query:"sort"           validate:"omitempty,oneof=asc desc"`
	Height         uint64      `query:"height"         validate:"omitempty,min=1"`
	Implementation string      `query:"implementation" validate:"omitempty,address"`
	Type           StringArray `query:"type"           validate:"omitempty,dive,proxy_contract_type"`
	Status         StringArray `query:"status"         validate:"omitempty,dive,proxy_contract_status"`
}

func (req *listProxyContracts) ToFilters(
	ctx context.Context,
	address storage.IAddress,
) (storage.ListProxyFilters, error) {
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}
	if req.Sort == "" {
		req.Sort = desc
	}
	filters := storage.ListProxyFilters{
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   pgSort(req.Sort),
	}

	if req.Height != 0 {
		filters.Height = pkgTypes.Level(req.Height)
	}

	if req.Implementation != "" {
		implementationHex, err := pkgTypes.HexFromString(req.Implementation)
		if err != nil {
			return filters, errors.Wrapf(err, "decoding proxy implementation address: %s", req.Implementation)
		}
		implementationAddress, err := address.ByHash(ctx, implementationHex)
		if err != nil {
			return filters, errors.Wrapf(err, "fetching implementation address by hash: %x", implementationAddress)
		}
		filters.ImplementationId = implementationAddress.Id
	}
	if len(req.Type) > 0 {
		filters.Type = make([]types.ProxyType, len(req.Type))
		for i := range req.Type {
			filters.Type[i] = types.ProxyType(req.Type[i])
		}
	}

	if len(req.Status) > 0 {
		filters.Status = make([]types.ProxyStatus, len(req.Status))
		for i := range req.Status {
			filters.Status[i] = types.ProxyStatus(req.Status[i])
		}
	}

	return filters, nil
}

// List godoc
//
//	@Summary		Proxy contracts list
//	@Description	List of all indexed proxy contracts
//	@Tags			proxy-contracts
//	@ID				list-proxy-contracts
//	@Param			limit			query	integer	false	"Count of requested entities"			 minimum(1)	maximum(100)
//	@Param			offset			query	integer	false	"Offset"								 minimum(0)
//	@Param			sort			query	string	false	"Sort order. Default: desc"				 Enums(asc, desc)
//	@Param			height			query	integer	false	"Proxy contract deploy height"			 minimum(1)
//	@Param			implementation	query	string	false	"Proxy contract implementation address"	 minlength(42)	maxlength(42)
//	@Param			type			query	string	false	"Comma-separated proxy contracts type"	 Enums(EIP1167, EIP7760, EIP7702, EIP1967, custom, clone_with_immutable_args)
//	@Param			status			query	string	false	"Comma-separated proxy contracts status" Enums(new, resolved, error)
//	@Produce		json
//	@Success		200	{array}		responses.ProxyContract
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/proxy-contracts [get]
func (handler *ProxyContractHandler) List(c echo.Context) error {
	req, err := bindAndValidate[listProxyContracts](c)
	if err != nil {
		return badRequestError(c, err)
	}
	filters, err := req.ToFilters(c.Request().Context(), handler.addresses)
	if err != nil {
		return badRequestError(c, err)
	}

	contracts, err := handler.contracts.FilteredList(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.contracts)
	}

	response := make([]responses.ProxyContract, len(contracts))
	for i := range contracts {
		response[i] = responses.NewProxyContract(contracts[i])
	}

	return returnArray(c, response)
}

package handler

import (
	"context"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/cmd/api/helpers"
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
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
	SortBy         string      `query:"sort_by"        validate:"omitempty,oneof=id height"`
	Height         uint64      `query:"height"         validate:"omitempty,min=1"`
	Implementation string      `query:"implementation" validate:"omitempty,address"`
	Type           StringArray `query:"type"           validate:"omitempty,dive,proxy_contract_type"`
	Status         StringArray `query:"status"         validate:"omitempty,dive,proxy_contract_status"`
	Cursor         string      `query:"cursor"         validate:"omitempty"`
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
		Limit:     req.Limit,
		Offset:    req.Offset,
		Sort:      pgSort(req.Sort),
		SortField: req.SortBy,
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
			return filters, errors.Wrapf(err, "fetching implementation address by hash: %x", implementationHex)
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

	if req.Cursor != "" {
		if req.SortBy != "" && req.SortBy != "id" {
			return filters, errCursorWithSortBy
		}
		if req.Offset > 0 {
			return filters, errCursorWithOffset
		}
		cursorID, err := helpers.DecodeIDCursor(req.Cursor)
		if err != nil {
			return filters, err
		}
		filters.CursorID = cursorID
	}

	return filters, nil
}

// List godoc
//
//	@Summary		List proxy contracts
//	@Description	Returns a paginated list of proxy contracts. Proxy contracts are smart contracts that delegate calls to implementation contracts. Can be filtered by type, status, implementation address, or deployment height.
//	@Tags			proxy-contracts
//	@ID				list-proxy-contracts
//	@Param			limit			query	integer	false	"Number of proxy contracts to return (default: 10)"																					minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of proxy contracts to skip (default: 0)"																					minimum(0)	default(0)
//	@Param			sort			query	string	false	"Sort order by deployment height (default: desc)"																					Enums(asc, desc)	default(desc)
//	@Param			sort_by			query	string	false	"Field to sort by (default: id)"																									Enums(id, height)
//	@Param			height			query	integer	false	"Filter by deployment block height"																									minimum(1)	example(12345)
//	@Param			implementation	query	string	false	"Filter by implementation contract address"																							minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			type			query	string	false	"Filter by proxy pattern (comma-separated list)"																					Enums(EIP1167, EIP7760, EIP7702, EIP1967, custom, clone_with_immutable_args)
//	@Param			status			query	string	false	"Filter by resolution status: new (just detected), resolved (implementation found), error (failed to resolve) (comma-separated)"	Enums(new, resolved, error)
//	@Param			cursor			query	string	false	"Opaque cursor for keyset pagination. Base64url-encoded value from the previous response's 'cursor' field. Encodes the id of the last returned record. Cannot be used together with offset (returns 400). Only supported when sort_by=id (default); returns 400 for other sort_by values."
//	@Produce		json
//	@Success		200	{object}	CursorResponse	"List of proxy contracts"
//	@Failure		400	{object}	Error					"Invalid request parameters"
//	@Failure		500	{object}	Error					"Internal server error"
//	@Router			/proxy [get]
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

	var cursor string
	if len(contracts) > 0 && (req.SortBy == "" || req.SortBy == "id") {
		last := contracts[len(contracts)-1]
		cursor = helpers.EncodeIDCursor(last.Id)
	}

	return returnCursorList(c, response, cursor)
}

package handler

import (
	"time"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
)

type UserOpHandler struct {
	userOps storage.IERC4337UserOps
	tx      storage.ITx
	address storage.IAddress
}

func NewUserOpHandler(
	userOps storage.IERC4337UserOps,
	tx storage.ITx,
	address storage.IAddress,
) *UserOpHandler {
	return &UserOpHandler{
		userOps: userOps,
		tx:      tx,
		address: address,
	}
}

type userOpListRequest struct {
	Limit     int     `query:"limit"     validate:"omitempty,min=1,max=100"`
	Offset    int     `query:"offset"    validate:"omitempty,min=0"`
	Sort      string  `query:"sort"      validate:"omitempty,oneof=asc desc"`
	Height    *uint64 `query:"height"    validate:"omitempty,min=0"`
	TxHash    string  `query:"tx"        validate:"omitempty,tx_hash"`
	Bundler   string  `query:"bundler"   validate:"omitempty,address"`
	Paymaster string  `query:"paymaster" validate:"omitempty,address"`
	Success   *bool   `query:"success"   validate:"omitempty"`

	From int64 `example:"1692892095" query:"from" swaggertype:"integer" validate:"omitempty,min=1"`
	To   int64 `example:"1692892095" query:"to"   swaggertype:"integer" validate:"omitempty,min=1"`
}

func (req *userOpListRequest) SetDefault() {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Sort == "" {
		req.Sort = desc
	}
}

// List godoc
//
//	@Summary		List ERC-4337 user operations
//	@Description	Returns a paginated list of ERC-4337 user operations. Can be filtered by transaction, block height, time range, bundler, paymaster, or success status.
//	@Tags			user_ops
//	@ID				list-user-ops
//	@Param			tx			query	string	false	"Filter by transaction hash (hexadecimal with 0x prefix)"	minlength(66)	maxlength(66)
//	@Param			limit		query	integer	false	"Number of user operations to return (default: 10)"			minimum(1)		maximum(100)	default(10)
//	@Param			offset		query	integer	false	"Number of user operations to skip (default: 0)"			minimum(0)		default(0)
//	@Param			height		query	integer	false	"Filter by block height"									minimum(0)
//	@Param			bundler		query	string	false	"Filter by bundler address"									minlength(42)	maxlength(42)
//	@Param			paymaster	query	string	false	"Filter by paymaster address"								minlength(42)	maxlength(42)
//	@Param			success		query	boolean	false	"Filter by success status"
//	@Param			from		query	integer	false	"Filter by timestamp from (Unix timestamp)"					minimum(1)
//	@Param			to			query	integer	false	"Filter by timestamp to (Unix timestamp)"					minimum(1)
//	@Param			sort		query	string	false	"Sort order by timestamp (default: desc)"					Enums(asc, desc)	default(desc)
//	@Produce		json
//	@Success		200	{array}		responses.UserOp	"List of user operations"
//	@Failure		400	{object}	Error				"Invalid request parameters"
//	@Failure		500	{object}	Error				"Internal server error"
//	@Router			/user_ops [get]
func (handler *UserOpHandler) List(c echo.Context) error {
	req, err := bindAndValidate[userOpListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filters := storage.ERC4337UserOpsListFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   pgSort(req.Sort),
	}

	if req.Height != nil {
		filters.Height = req.Height
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

	if req.Bundler != "" {
		hash, err := types.HexFromString(req.Bundler)
		if err != nil {
			return badRequestError(c, err)
		}

		address, err := handler.address.ByHash(c.Request().Context(), hash)
		if err != nil {
			return handleError(c, err, handler.address)
		}

		filters.BundlerId = &address.Id
	}

	if req.Paymaster != "" {
		hash, err := types.HexFromString(req.Paymaster)
		if err != nil {
			return badRequestError(c, err)
		}

		address, err := handler.address.ByHash(c.Request().Context(), hash)
		if err != nil {
			return handleError(c, err, handler.address)
		}

		filters.PaymasterId = &address.Id
	}

	if req.Success != nil {
		filters.Success = req.Success
	}

	if req.From > 0 {
		filters.TimeFrom = time.Unix(req.From, 0).UTC()
	}
	if req.To > 0 {
		filters.TimeTo = time.Unix(req.To, 0).UTC()
	}

	userOps, err := handler.userOps.Filter(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.userOps)
	}

	response := make([]responses.UserOp, len(userOps))
	for i := range userOps {
		response[i] = responses.NewUserOp(userOps[i])
	}

	return returnArray(c, response)
}

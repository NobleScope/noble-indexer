package handler

import (
	"net/http"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
)

type BeaconWithdrawalHandler struct {
	address           storage.IAddress
	beaconWithdrawals storage.IBeaconWithdrawal
}

func NewBeaconWithdrawalHandler(
	beaconWithdrawals storage.IBeaconWithdrawal,
	address storage.IAddress,
) *BeaconWithdrawalHandler {
	return &BeaconWithdrawalHandler{
		beaconWithdrawals: beaconWithdrawals,
		address:           address,
	}
}

type beaconWithdrawalListRequest struct {
	Limit   int          `query:"limit"   validate:"omitempty,min=1,max=100"`
	Offset  int          `query:"offset"  validate:"omitempty,min=0"`
	Sort    string       `query:"sort"    validate:"omitempty,oneof=asc desc"`
	Height  *types.Level `query:"height"  validate:"omitempty,min=0"`
	Address string       `query:"address" validate:"omitempty,address"`
}

func (p *beaconWithdrawalListRequest) SetDefault() {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Sort == "" {
		p.Sort = asc
	}
}

// List godoc
//
//	@Summary		List beacon chain withdrawals
//	@Description	Returns a paginated list of beacon chain (consensus layer) withdrawals. Withdrawals represent ETH transferred from validators to execution layer addresses. Can be filtered by block height or recipient address.
//	@Tags			beacon
//	@ID				list-beacon-withdrawals
//	@Param			limit	query	integer	false	"Number of withdrawals to return (default: 10)"			minimum(1)	maximum(100)	default(10)
//	@Param			offset	query	integer	false	"Number of withdrawals to skip (default: 0)"			minimum(0)	default(0)
//	@Param			sort	query	string	false	"Sort order by block height (default: asc)"				Enums(asc, desc)	default(asc)
//	@Param			height	query	integer	false	"Filter by block height"								minimum(0)	example(12345)
//	@Param			address	query	string	false	"Filter by recipient address (hexadecimal with 0x prefix)"	minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Produce		json
//	@Success		200	{array}		responses.BeaconWithdrawal	"List of beacon withdrawals"
//	@Failure		400	{object}	Error						"Invalid request parameters"
//	@Failure		500	{object}	Error						"Internal server error"
//	@Router			/beacon_withdrawals [get]
func (handler *BeaconWithdrawalHandler) List(c echo.Context) error {
	req, err := bindAndValidate[beaconWithdrawalListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filter := storage.BeaconWithdrawalListFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Height: req.Height,
		Sort:   pgSort(req.Sort),
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
		filter.AddressId = &address.Id
	}
	withdrawals, err := handler.beaconWithdrawals.Filter(c.Request().Context(), filter)
	if err != nil {
		return handleError(c, err, handler.address)
	}

	response := make([]responses.BeaconWithdrawal, len(withdrawals))
	for i := range withdrawals {
		response[i] = responses.NewBeaconWithdrawal(withdrawals[i])
	}

	return c.JSON(http.StatusOK, response)
}

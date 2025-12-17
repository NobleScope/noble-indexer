package handler

import (
	"context"
	"strconv"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type SearchHandler struct {
	search  storage.ISearch
	address storage.IAddress
	block   storage.IBlock
	tx      storage.ITx
	token   storage.IToken
}

func NewSearchHandler(
	search storage.ISearch,
	address storage.IAddress,
	block storage.IBlock,
	tx storage.ITx,
	token storage.IToken,
) SearchHandler {
	return SearchHandler{
		search:  search,
		address: address,
		block:   block,
		tx:      tx,
		token:   token,
	}
}

type searchRequest struct {
	Search string `query:"query" validate:"required"`
}

// Search godoc
//
//	@Summary				Search by hash
//	@Description			search
//	@Tags					search
//	@ID						search
//	@Param					query	query	string	true	"Search string"
//	@Produce				json
//	@Success				200	{array}	responses.SearchItem
//	@Success				204
//	@Failure				400	{object}	Error
//	@Failure				500	{object}	Error
//	@Router					/search [get]
func (handler SearchHandler) Search(c echo.Context) error {
	req, err := bindAndValidate[searchRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	data := make([]responses.SearchItem, 0)

	if height, err := strconv.ParseInt(req.Search, 10, 64); err == nil {
		block, err := handler.block.ByHeight(c.Request().Context(), types.Level(height), false)
		if err == nil {
			data = append(data, responses.SearchItem{
				Type:   "block",
				Result: responses.NewBlock(block),
			})
		}
	}

	var response []responses.SearchItem

	switch {
	case evmAddressRegex.MatchString(req.Search):
		response, err = handler.searchAddress(c.Request().Context(), req.Search)
	case evmTransactionHashRegex.MatchString(req.Search):
		response, err = handler.searchHash(c.Request().Context(), req.Search)
	default:
		response, err = handler.searchText(c.Request().Context(), req.Search)
	}
	if err != nil {
		if !handler.address.IsNoRows(err) {
			return handleError(c, err, handler.address)
		}
	}

	data = append(data, response...)
	return returnArray(c, data)
}

func (handler SearchHandler) searchAddress(ctx context.Context, search string) ([]responses.SearchItem, error) {
	hash, err := types.HexFromString(search)
	if err != nil {
		return nil, err
	}

	address, err := handler.address.ByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	result := responses.SearchItem{
		Type:   "address",
		Result: responses.NewAddress(address),
	}
	return []responses.SearchItem{result}, nil
}

func (handler SearchHandler) searchHash(ctx context.Context, search string) ([]responses.SearchItem, error) {
	hash, err := types.HexFromString(search)
	if err != nil {
		return nil, err
	}

	result, err := handler.search.Search(ctx, hash)
	if err != nil {
		return nil, err
	}

	response := make([]responses.SearchItem, len(result))
	for i := range result {
		response[i].Type = result[i].Type
		switch response[i].Type {
		case "tx":
			tx, err := handler.tx.GetByID(ctx, result[i].Id)
			if err != nil {
				return nil, err
			}
			response[i].Result = responses.NewTransaction(*tx)
		case "block":
			block, err := handler.block.GetByID(ctx, result[i].Id)
			if err != nil {
				return nil, err
			}
			response[i].Result = responses.NewBlock(*block)
		}
	}

	return response, nil
}

func (handler SearchHandler) searchText(ctx context.Context, text string) ([]responses.SearchItem, error) {
	result, err := handler.search.SearchText(ctx, text)
	if err != nil {
		return nil, err
	}

	response := make([]responses.SearchItem, len(result))
	for i := range result {
		response[i].Type = result[i].Type
		switch response[i].Type {
		case "token":
			token, err := handler.token.GetByID(ctx, result[i].Id)
			if err != nil {
				return nil, err
			}
			response[i].Result = responses.NewToken(*token)
		default:
			return nil, errors.Errorf("unknown search text type: %s", response[i].Type)
		}
	}

	return response, nil
}

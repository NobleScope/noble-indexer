package handler

import (
	"net/http"
	"time"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	internalTypes "github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

type TokenHandler struct {
	token    storage.IToken
	transfer storage.ITransfer
	tbs      storage.ITokenBalance
	address  storage.IAddress
	tx       storage.ITx
}

func NewTokenHandler(
	token storage.IToken,
	transfer storage.ITransfer,
	tbs storage.ITokenBalance,
	address storage.IAddress,
	tx storage.ITx,
) *TokenHandler {
	return &TokenHandler{
		token:    token,
		transfer: transfer,
		tbs:      tbs,
		address:  address,
		tx:       tx,
	}
}

type tokenListRequest struct {
	Contract string      `query:"contract" validate:"omitempty,address"`
	Limit    int         `query:"limit"    validate:"omitempty,min=1,max=100"`
	Offset   int         `query:"offset"   validate:"omitempty,min=0"`
	Type     StringArray `query:"type"     validate:"omitempty,dive,token_type"`
	Sort     string      `query:"sort"     validate:"omitempty,oneof=asc desc"`
}

func (req *tokenListRequest) SetDefault() {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Sort == "" {
		req.Sort = desc
	}
}

// List godoc
//
//	@Summary		List tokens
//	@Description	List tokens
//	@Tags			token
//	@ID				list-tokens
//	@Param			contract		query	string	false	"Contract address which issued the token"			minlength(42)	maxlength(42)
//	@Param			limit			query	integer	false	"Count of requested entities"						minimum(1)	maximum(100)
//	@Param			offset			query	integer	false	"Offset"											minimum(0)
//	@Param			type			query	string	false	"Comma-separated list of token types"				Enums(ERC20, ERC721, ERC1155)
//	@Param			sort			query	string	false	"Sort order. Default: desc"							Enums(asc, desc)
//	@Produce		json
//	@Success		200	{array}		responses.Token
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/tokens [get]
func (handler *TokenHandler) List(c echo.Context) error {
	req, err := bindAndValidate[tokenListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	tokenTypes := make([]internalTypes.TokenType, len(req.Type))
	for i := range tokenTypes {
		tokenTypes[i] = internalTypes.TokenType(req.Type[i])
	}

	filters := storage.TokenListFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   pgSort(req.Sort),
		Type:   tokenTypes,
	}
	if req.Contract != "" {
		address, err := handler.getAddressByHash(c, req.Contract)
		if err != nil {
			return err
		}
		filters.ContractId = &address.Id
	}

	tokens, err := handler.token.Filter(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.token)
	}

	response := make([]responses.Token, len(tokens))
	for i := range tokens {
		response[i] = responses.NewToken(tokens[i])
	}

	return returnArray(c, response)
}

type tokenRequest struct {
	Contract string `param:"contract" validate:"required,address"`
	TokenId  string `param:"token_id" validate:"required,min=0"`
}

// Get godoc
//
//	@Summary		Get token info
//	@Description	Get token info
//	@Tags			token
//	@ID				get-token
//	@Param			contract	path	string	true	"Contract address"	minlength(42)	maxlength(42)
//	@Param			token_id	path	string	true	"Token ID"
//	@Produce		json
//	@Success		200	{object}	responses.Token
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/tokens/{contract}/{token_id} [get]
func (handler *TokenHandler) Get(c echo.Context) error {
	req, err := bindAndValidate[tokenRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	hash, err := types.HexFromString(req.Contract)
	if err != nil {
		return badRequestError(c, err)
	}

	address, err := handler.address.ByHash(c.Request().Context(), hash)
	if err != nil {
		return handleError(c, err, handler.address)
	}

	tokenId, err := decimal.NewFromString(req.TokenId)
	if err != nil {
		return badRequestError(c, err)
	}
	token, err := handler.token.Get(c.Request().Context(), address.Id, tokenId)
	if err != nil {
		return handleError(c, err, handler.token)
	}

	return c.JSON(http.StatusOK, responses.NewToken(token))
}

type transferListRequest struct {
	Limit       int         `query:"limit"        validate:"omitempty,min=1,max=100"`
	Offset      int         `query:"offset"       validate:"omitempty,min=0"`
	Sort        string      `query:"sort"         validate:"omitempty,oneof=asc desc"`
	Height      *uint64     `query:"height"       validate:"omitempty,min=0"`
	TxHash      string      `query:"tx_hash"      validate:"omitempty,tx_hash"`
	Type        StringArray `query:"type"         validate:"omitempty"`
	AddressFrom string      `query:"address_from" validate:"omitempty,address"`
	AddressTo   string      `query:"address_to"   validate:"omitempty,address"`
	Contract    string      `query:"contract"     validate:"omitempty,address"`
	TokenId     *string     `query:"token_id"     validate:"omitempty"`

	From int64 `example:"1692892095" query:"time_from" swaggertype:"integer" validate:"omitempty,min=1,max=16725214800"`
	To   int64 `example:"1692892095" query:"time_to"   swaggertype:"integer" validate:"omitempty,min=1,max=16725214800"`
}

func (p *transferListRequest) SetDefault() {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Sort == "" {
		p.Sort = asc
	}
}

// TransferList godoc
//
//	@Summary		List token transfers
//	@Description	List token transfers
//	@Tags			token
//	@ID				list-token-transfers
//	@Param			limit			query	integer	false	"Count of requested entities"						minimum(1)	maximum(100)
//	@Param			offset			query	integer	false	"Offset"											minimum(0)
//	@Param			sort			query	string	false	"Sort order. Default: desc"							Enums(asc, desc)
//	@Param			height			query	integer	false	"Block height"										minimum(0)
//	@Param			time_from		query	integer	false	"Time from in unix timestamp"						mininum(1)  maximum(2147483647)
//	@Param			time_to			query	integer	false	"Time to in unix timestamp"							mininum(1)  maximum(2147483647)
//	@Param			type			query	string	false	"Comma-separated list of token types"				Enums(burn, mint, transfer, unknown)
//	@Param			tx_hash			query	string	false	"Transaction hash in hexadecimal with 0x prefix"	minlength(66)	maxlength(66)
//	@Param			address_from	query	string	false	"Address from"										minlength(42)	maxlength(42)
//	@Param			address_to		query	string	false	"Address to"										minlength(42)	maxlength(42)
//	@Param			contract		query	string	false	"Contract address"									minlength(42)	maxlength(42)
//	@Param			token_id		query	string	false	"Token ID"
//	@Produce		json
//	@Success		200	{array}		responses.Transfer
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/transfers [get]
func (handler *TokenHandler) TransferList(c echo.Context) error {
	req, err := bindAndValidate[transferListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	transferTypes := make([]internalTypes.TransferType, len(req.Type))
	for i := range transferTypes {
		transferTypes[i] = internalTypes.TransferType(req.Type[i])
	}

	filters := storage.TransferListFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   pgSort(req.Sort),
		Height: req.Height,
		Type:   transferTypes,
	}

	if req.TokenId != nil {
		tokenId, err := decimal.NewFromString(*req.TokenId)
		if err != nil {
			return badRequestError(c, err)
		}
		filters.TokenId = &tokenId
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

	if req.Contract != "" {
		address, err := handler.getAddressByHash(c, req.Contract)
		if err != nil {
			return err
		}
		filters.ContractId = &address.Id
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

	if req.Height != nil {
		filters.Height = req.Height
	}

	if req.From > 0 {
		filters.TimeFrom = time.Unix(req.From, 0).UTC()
	}

	if req.To > 0 {
		filters.TimeTo = time.Unix(req.To, 0).UTC()
	}

	transfers, err := handler.transfer.Filter(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.token)
	}

	response := make([]responses.Transfer, len(transfers))
	for i := range transfers {
		response[i] = responses.NewTransfer(transfers[i])
	}

	return returnArray(c, response)
}

type tokenTransferRequest struct {
	Id uint64 `param:"id" validate:"required,min=1"`
}

// GetTransfer godoc
//
//	@Summary		Get token transfer info
//	@Description	Get token transfer info
//	@Tags			token
//	@ID				get-token-transfer
//	@Param			contract	path	integer	true	"Internal id"	mininum(1)
//	@Produce		json
//	@Success		200	{object}	responses.Transfer
//	@Success		204
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/transfers/{id} [get]
func (handler *TokenHandler) GetTransfer(c echo.Context) error {
	req, err := bindAndValidate[tokenTransferRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}

	transfer, err := handler.transfer.Get(c.Request().Context(), req.Id)
	if err != nil {
		return handleError(c, err, handler.transfer)
	}

	return c.JSON(http.StatusOK, responses.NewTransfer(transfer))
}

type tokenBalanceListRequest struct {
	Limit    int     `query:"limit"    validate:"omitempty,min=1,max=100"`
	Offset   int     `query:"offset"   validate:"omitempty,min=0"`
	Sort     string  `query:"sort"     validate:"omitempty,oneof=asc desc"`
	Address  string  `query:"address"  validate:"omitempty,address"`
	Contract string  `query:"contract" validate:"omitempty,address"`
	TokenId  *string `query:"token_id" validate:"omitempty"`
}

func (p *tokenBalanceListRequest) SetDefault() {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Sort == "" {
		p.Sort = asc
	}
}

// TokenBalanceList godoc
//
//	@Summary		List token balances
//	@Description	List token balances
//	@Tags			token
//	@ID				list-token-balances
//	@Param			limit			query	integer	false	"Count of requested entities"						minimum(1)	maximum(100)
//	@Param			offset			query	integer	false	"Offset"											minimum(0)
//	@Param			sort			query	string	false	"Sort order. Default: desc"							Enums(asc, desc)
//	@Param			address			query	string	false	"Address"											minlength(42)	maxlength(42)
//	@Param			contract		query	string	false	"Contract address"									minlength(42)	maxlength(42)
//	@Param			token_id		query	string	false	"Token ID"
//	@Produce		json
//	@Success		200	{array}		responses.TokenBalance
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/token_balances [get]
func (handler *TokenHandler) TokenBalanceList(c echo.Context) error {
	req, err := bindAndValidate[tokenBalanceListRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	req.SetDefault()

	filters := storage.TokenBalanceListFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   pgSort(req.Sort),
	}

	if req.TokenId != nil {
		tokenId, err := decimal.NewFromString(*req.TokenId)
		if err != nil {
			return badRequestError(c, err)
		}
		filters.TokenId = &tokenId
	}

	if req.Contract != "" {
		address, err := handler.getAddressByHash(c, req.Contract)
		if err != nil {
			return err
		}
		filters.ContractId = &address.Id
	}

	if req.Address != "" {
		address, err := handler.getAddressByHash(c, req.Address)
		if err != nil {
			return err
		}
		filters.AddressId = &address.Id
	}

	tbs, err := handler.tbs.Filter(c.Request().Context(), filters)
	if err != nil {
		return handleError(c, err, handler.token)
	}

	response := make([]responses.TokenBalance, len(tbs))
	for i := range tbs {
		response[i] = responses.NewTokenBalance(tbs[i])
	}

	return returnArray(c, response)
}

func (handler *TokenHandler) getAddressByHash(c echo.Context, h string) (storage.Address, error) {
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

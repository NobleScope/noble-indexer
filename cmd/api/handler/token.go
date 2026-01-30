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
//	@Description	Returns a paginated list of tokens (ERC20, ERC721, ERC1155). Can be filtered by token type or issuing contract address.
//	@Tags			token
//	@ID				list-tokens
//	@Param			contract		query	string	false	"Filter by contract address that issued the token"	minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			limit			query	integer	false	"Number of tokens to return (default: 10)"			minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of tokens to skip (default: 0)"				minimum(0)	default(0)
//	@Param			type			query	string	false	"Filter by token standard (comma-separated list)"	Enums(ERC20, ERC721, ERC1155)
//	@Param			sort			query	string	false	"Sort order by creation time (default: desc)"		Enums(asc, desc)	default(desc)
//	@Produce		json
//	@Success		200	{array}		responses.Token	"List of tokens"
//	@Failure		400	{object}	Error			"Invalid request parameters"
//	@Failure		500	{object}	Error			"Internal server error"
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
//	@Summary		Get token by contract and ID
//	@Description	Returns detailed information about a specific token including metadata, supply, and holder information. For ERC20 tokens use token_id=0, for ERC721/ERC1155 use the specific token ID.
//	@Tags			token
//	@ID				get-token
//	@Param			contract	path	string	true	"Token contract address"		minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			token_id	path	string	true	"Token ID (use 0 for ERC20)"	example(0)
//	@Produce		json
//	@Success		200	{object}	responses.Token	"Token information"
//	@Success		204								"Token not found"
//	@Failure		400	{object}	Error			"Invalid contract address or token ID"
//	@Failure		500	{object}	Error			"Internal server error"
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
	Type        StringArray `query:"type"         validate:"omitempty,dive,transfer_type"`
	AddressFrom string      `query:"address_from" validate:"omitempty,address"`
	AddressTo   string      `query:"address_to"   validate:"omitempty,address"`
	Contract    string      `query:"contract"     validate:"omitempty,address"`
	TokenId     *string     `query:"token_id"     validate:"omitempty"`

	From int64 `example:"1692892095" query:"time_from" swaggertype:"integer" validate:"omitempty,min=1"`
	To   int64 `example:"1692892095" query:"time_to"   swaggertype:"integer" validate:"omitempty,min=1"`
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
//	@Description	Returns a paginated list of token transfer events (mint, burn, transfer). Can be filtered by transfer type, addresses, contract, token ID, transaction, block height, or time range.
//	@Tags			token
//	@ID				list-token-transfers
//	@Param			limit			query	integer	false	"Number of transfers to return (default: 10)"				minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of transfers to skip (default: 0)"					minimum(0)	default(0)
//	@Param			sort			query	string	false	"Sort order by timestamp (default: asc)"					Enums(asc, desc)	default(asc)
//	@Param			height			query	integer	false	"Filter by block height"									minimum(0)	example(12345)
//	@Param			time_from		query	integer	false	"Filter by timestamp from (Unix timestamp)"					minimum(1)	example(1692892095)
//	@Param			time_to			query	integer	false	"Filter by timestamp to (Unix timestamp)"					minimum(1)	example(1692892095)
//	@Param			type			query	string	false	"Filter by transfer type (comma-separated list)"			Enums(burn, mint, transfer, unknown)
//	@Param			tx_hash			query	string	false	"Filter by transaction hash (hexadecimal with 0x prefix)"	minlength(66)	maxlength(66)	example(0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef)
//	@Param			address_from	query	string	false	"Filter by sender address"									minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			address_to		query	string	false	"Filter by recipient address"								minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			contract		query	string	false	"Filter by token contract address"							minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			token_id		query	string	false	"Filter by token ID"										example(0)
//	@Produce		json
//	@Success		200	{array}		responses.Transfer	"List of token transfers"
//	@Failure		400	{object}	Error				"Invalid request parameters"
//	@Failure		500	{object}	Error				"Internal server error"
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
//	@Summary		Get token transfer by ID
//	@Description	Returns detailed information about a specific token transfer event by its internal database ID
//	@Tags			token
//	@ID				get-token-transfer
//	@Param			id	path	integer	true	"Transfer internal ID"	minimum(1)	example(12345)
//	@Produce		json
//	@Success		200	{object}	responses.Transfer	"Transfer information"
//	@Success		204									"Transfer not found"
//	@Failure		400	{object}	Error				"Invalid transfer ID"
//	@Failure		500	{object}	Error				"Internal server error"
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
//	@Description	Returns a paginated list of token balances showing how many tokens each address holds. Can be filtered by address, contract, or token ID. Useful for finding token holders.
//	@Tags			token
//	@ID				list-token-balances
//	@Param			limit			query	integer	false	"Number of balances to return (default: 10)"	minimum(1)	maximum(100)	default(10)
//	@Param			offset			query	integer	false	"Number of balances to skip (default: 0)"		minimum(0)	default(0)
//	@Param			sort			query	string	false	"Sort order (default: asc)"						Enums(asc, desc)	default(asc)
//	@Param			address			query	string	false	"Filter by holder address"						minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			contract		query	string	false	"Filter by token contract address"				minlength(42)	maxlength(42)	example(0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb)
//	@Param			token_id		query	string	false	"Filter by token ID"							example(0)
//	@Produce		json
//	@Success		200	{array}		responses.TokenBalance	"List of token balances"
//	@Failure		400	{object}	Error					"Invalid request parameters"
//	@Failure		500	{object}	Error					"Internal server error"
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

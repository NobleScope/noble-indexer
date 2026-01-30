package responses

import (
	"encoding/json"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
)

// Token model info
//
//	@Description	Noble token information
type Token struct {
	Id             uint64 `example:"0"                                          json:"id,omitempty"     swaggertype:"integer"`
	Contract       string `example:"0xdAC17F958D2ee523a2206206994597C13D831ec7" json:"contract"         swaggertype:"string"`
	TokenId        string `example:"0"                                          json:"token_id"         swaggertype:"string"`
	Type           string `example:"ERC20"                                      json:"type"             swaggertype:"string"`
	Name           string `example:"Tether USD"                                 json:"name,omitempty"   swaggertype:"string"`
	Symbol         string `example:"USD"                                        json:"symbol,omitempty" swaggertype:"string"`
	Decimals       uint8  `example:"6"                                          json:"decimals"         swaggertype:"integer"`
	TransfersCount uint64 `example:"123"                                        json:"transfers_count"  swaggertype:"integer"`
	Supply         string `example:"123456789"                                  json:"supply"           swaggertype:"string"`
	Logo           string `example:"http://site.com/image.png"                  json:"logo,omitempty"   swaggertype:"string"`

	Metadata json.RawMessage `json:"metadata,omitempty"`
}

func NewToken(token storage.Token) Token {
	t := Token{
		Id:             token.Id,
		Contract:       token.Contract.Address.Hash.Hex(),
		TokenId:        token.TokenID.String(),
		Type:           token.Type.String(),
		Name:           token.Name,
		Symbol:         token.Symbol,
		Decimals:       token.Decimals,
		TransfersCount: token.TransfersCount,
		Supply:         token.Supply.String(),
		Metadata:       token.Metadata,
		Logo:           token.Logo,
	}

	return t
}

// Transfer model info
//
//	@Description	Token transfer information
type Transfer struct {
	Id     uint64    `example:"0"                                                                  json:"id"          swaggertype:"integer"`
	Height uint64    `example:"100"                                                                json:"height"      swaggertype:"integer"`
	Time   time.Time `example:"2026-01-01T01:01:01+00:00"                                          format:"date-time" json:"time"           swaggertype:"string"`
	Amount string    `example:"123456789"                                                          json:"amount"      swaggertype:"string"`
	Type   string    `example:"transfer"                                                           json:"type"        swaggertype:"string"`
	From   string    `example:"0x0000000000000000000000000000000000000001"                         json:"from"        swaggertype:"string"`
	To     string    `example:"0x0000000000000000000000000000000000000002"                         json:"to"          swaggertype:"string"`
	TxHash string    `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"tx_hash"     swaggertype:"string"`

	Token *Token `json:"token,omitempty"`
}

func NewTransfer(transfer storage.Transfer) Transfer {
	t := Transfer{
		Id:     transfer.Id,
		Height: uint64(transfer.Height),
		Time:   transfer.Time,
		Amount: transfer.Amount.String(),
		Type:   transfer.Type.String(),
		TxHash: transfer.Tx.Hash.Hex(),
	}

	if transfer.FromAddress != nil {
		t.From = transfer.FromAddress.Hash.Hex()
	}
	if transfer.ToAddress != nil {
		t.To = transfer.ToAddress.Hash.Hex()
	}
	if transfer.Token != nil {
		token := NewToken(*transfer.Token)
		token.Contract = transfer.Contract.Address.Hash.Hex()
		token.TokenId = transfer.TokenID.String()
		t.Token = &token
	}

	return t
}

// TokenBalance model info
//
//	@Description	Token balance information
type TokenBalance struct {
	Address string `example:"0x0000000000000000000000000000000000000001" json:"address" swaggertype:"string"`
	Value   string `example:"123456789"                                  json:"value"   swaggertype:"string"`

	Token Token `json:"token"`
}

func NewTokenBalance(tb storage.TokenBalance) TokenBalance {
	t := TokenBalance{
		Address: tb.Address.Hash.Hex(),
		Value:   tb.Balance.String(),
		Token:   NewToken(tb.Token),
	}
	t.Token.Contract = tb.Contract.Address.Hash.Hex()
	t.Token.TokenId = tb.TokenID.String()
	return t
}

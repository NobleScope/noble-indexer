package responses

import (
	"github.com/baking-bad/noble-indexer/internal/storage/types"
)

// Params model info
//
//	@Description	API request parameters
type Params map[string]string

// Enums model info
//
//	@Description	Available enum values for various entity types
type Enums struct {
	TokenType    []string `example:"ERC20,ERC721,ERC1155" json:"token_type"    swaggertype:"array,string"`
	TraceType    []string `example:"call,create"          json:"trace_type"    swaggertype:"array,string"`
	TransferType []string `example:"transfer,mint,burn"   json:"transfer_type" swaggertype:"array,string"`
	TxStatus     []string `example:"success,revert"       json:"tx_status"     swaggertype:"array,string"`
	TxType       []string `example:"legacy,dynamic_fee"   json:"tx_type"       swaggertype:"array,string"`
}

func NewEnums() Enums {
	return Enums{
		TokenType:    types.TokenTypeNames(),
		TraceType:    types.TraceTypeNames(),
		TransferType: types.TransferTypeNames(),
		TxStatus:     types.TxStatusNames(),
		TxType:       types.TxTypeNames(),
	}
}

package responses

import (
	"github.com/baking-bad/noble-indexer/internal/storage/types"
)

type Params map[string]string

type Enums struct {
	TokenType    []string `json:"token_type"`
	TraceType    []string `json:"trace_type"`
	TransferType []string `json:"transfer_type"`
	TxStatus     []string `json:"tx_status"`
	TxType       []string `json:"tx_type"`
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

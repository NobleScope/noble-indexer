package responses

import (
	"github.com/NobleScope/noble-indexer/internal/storage/types"
)

// Params model info
//
//	@Description	API request parameters
type Params map[string]string

// Enums model info
//
//	@Description	Available enum values for various entity types
type Enums struct {
	TokenType              []string `example:"ERC20,ERC721,ERC1155"       json:"token_type"               swaggertype:"array,string"`
	TraceType              []string `example:"call,create"                json:"trace_type"               swaggertype:"array,string"`
	CallType               []string `example:"call,delegatecall"          json:"call_type"                swaggertype:"array,string"`
	TransferType           []string `example:"transfer,mint,burn"         json:"transfer_type"            swaggertype:"array,string"`
	TxStatus               []string `example:"success,revert"             json:"tx_status"                swaggertype:"array,string"`
	TxType                 []string `example:"legacy,dynamic_fee"         json:"tx_type"                  swaggertype:"array,string"`
	ProxyType              []string `example:"eip1167,eip1967"            json:"proxy_type"               swaggertype:"array,string"`
	ProxyStatus            []string `example:"resolved,unresolved"        json:"proxy_status"             swaggertype:"array,string"`
	MetadataStatus         []string `example:"new,success,failed"         json:"metadata_status"          swaggertype:"array,string"`
	VerificationTaskStatus []string `example:"new,pending,success,failed" json:"verification_task_status" swaggertype:"array,string"`
	EVMVersion             []string `example:"shanghai,cancun,prague"     json:"evm_version"              swaggertype:"array,string"`
	LicenseType            []string `example:"mit,apache_2_0,gnu_gpl_v3"  json:"license_type"             swaggertype:"array,string"`
}

func NewEnums() Enums {
	return Enums{
		TokenType:              types.TokenTypeNames(),
		TraceType:              types.TraceTypeNames(),
		CallType:               types.CallTypeNames(),
		TransferType:           types.TransferTypeNames(),
		TxStatus:               types.TxStatusNames(),
		TxType:                 types.TxTypeNames(),
		ProxyType:              types.ProxyTypeNames(),
		ProxyStatus:            types.ProxyStatusNames(),
		MetadataStatus:         types.MetadataStatusNames(),
		VerificationTaskStatus: types.VerificationTaskStatusNames(),
		EVMVersion:             types.EVMVersionNames(),
		LicenseType:            types.LicenseTypeNames(),
	}
}

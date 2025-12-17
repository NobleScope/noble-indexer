package responses

import (
	"encoding/json"

	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
)

// Contract model info
//
//	@Description	Noble contract information
type Contract struct {
	Id               uint64 `example:"321"                                                                 json:"id"                          swaggertype:"integer"`
	Address          string `example:"0x0000000000000000000000000000000000000001"                          json:"address"                     swaggertype:"string"`
	Implementation   string `example:"0x0000000000000000000000000000000000000001"                          json:"implementation,omitempty"    swaggertype:"string"`
	Code             string `example:"0x01234567890123456789012345678901234567890123456789"                json:"code,omitempty"              swaggertype:"string"`
	Verified         bool   `example:"false"                                                               json:"verified"                    swaggertype:"boolean"`
	TxHash           string `example:"0x0000000000000000000000000000000000000002"                          json:"tx_hash"                     swaggertype:"string"`
	CompilerVersion  string `example:"0.1.1"                                                               json:"compiler_version,omitempty"  swaggertype:"string"`
	MetadataLink     string `example:"https://ipfs.io/ipfs/QmWYtNwHxXxzhrWj7TUcB5tC3m1bSGFXAtEqwegMbk1sjt" json:"metadata_link,omitempty"     swaggertype:"string"`
	OptimizerEnabled bool   `example:"true"                                                                json:"optimizer_enabled,omitempty" swaggertype:"boolean"`
	Language         string `example:"Solidity"                                                            json:"language,omitempty"          swaggertype:"string"`
	Error            string `example:"Error string"                                                        json:"error,omitempty"             swaggertype:"string"`

	Tags []string        `json:"tags,omitempty"`
	ABI  json.RawMessage `json:"abi,omitempty"`
}

func NewContract(contract storage.Contract) Contract {
	addressBytes, err := pkgTypes.HexFromString(contract.Address.String())
	if err != nil {
		panic(err)
	}
	c := Contract{
		Id:               contract.Id,
		Address:          addressBytes.Hex(),
		Code:             contract.Code.Hex(),
		Verified:         contract.Verified,
		CompilerVersion:  contract.CompilerVersion,
		MetadataLink:     contract.MetadataLink,
		OptimizerEnabled: contract.OptimizerEnabled,
		Language:         contract.Language,
		Error:            contract.Error,
		Tags:             contract.Tags,
		ABI:              contract.ABI,
	}

	if contract.Tx != nil {
		c.TxHash = contract.Tx.Hash.Hex()
	}

	if contract.Implementation != nil {
		implementationBytes, err := pkgTypes.HexFromString(*contract.Implementation)
		if err != nil {
			panic(err)
		}
		c.Implementation = implementationBytes.Hex()
	}

	return c
}

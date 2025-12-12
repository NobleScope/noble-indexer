package responses

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
)

type ProxyContract struct {
	Height         uint64  `example:"100"                                                            json:"height"         swaggertype:"integer"`
	Contract       string  `example:"0x0000000000000000000000000000000000000000"                     json:"contract"       swaggertype:"string"`
	Type           string  `enums:"EIP1167,EIP7760,EIP7702,EIP1967,custom,clone_with_immutable_args" example:"EIP1967"     json:"type"           swaggertype:"string"`
	Status         string  `enums:"new,resolved,error"                                               example:"resolved"    json:"status"         swaggertype:"string"`
	Implementation *string `example:"0x0000000000000000000000000000000000000001"                     json:"implementation" swaggertype:"string"`
}

func NewProxyContract(pc storage.ProxyContract) ProxyContract {
	contractAddress, err := pkgTypes.HexFromString(pc.Contract.Address.Address)
	if err != nil {
		panic(err)
	}
	result := ProxyContract{
		Height:   uint64(pc.Height),
		Contract: contractAddress.Hex(),
		Type:     string(pc.Type),
		Status:   string(pc.Status),
	}

	if pc.Implementation != nil {
		implAddress, err := pkgTypes.HexFromString(pc.Implementation.Address.Address)
		if err != nil {
			panic(err)
		}
		impl := implAddress.Hex()
		result.Implementation = &impl
	}

	return result
}

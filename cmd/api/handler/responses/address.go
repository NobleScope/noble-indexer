package responses

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
)

// Address model info
//
//	@Description	Noble address information
type Address struct {
	Id                uint64         `example:"321"                                        json:"id"                 swaggertype:"integer"`
	FirstHeight       pkgTypes.Level `example:"100"                                        json:"first_first_height" swaggertype:"integer"`
	LastHeight        pkgTypes.Level `example:"100"                                        json:"last_height"        swaggertype:"integer"`
	Hash              string         `example:"0xd90d69b7cf347b5bfe0719baf7eef310c085e46b" json:"hash"               swaggertype:"string"`
	IsContract        bool           `example:"false"                                      json:"is_contract"        swaggertype:"boolean"`
	TxCount           int            `example:"23456"                                      json:"txs_count"          swaggertype:"integer"`
	DeployedContracts int            `example:"7"                                          json:"deployed_contracts" swaggertype:"integer"`
	Interactions      int            `example:"890"                                        json:"interactions"       swaggertype:"integer"`
	Balance           Balance        `json:"balance"`
}

func NewAddress(addr storage.Address) Address {
	address := Address{
		Id:                addr.Id,
		FirstHeight:       addr.FirstHeight,
		LastHeight:        addr.LastHeight,
		Hash:              addr.Hash.Hex(),
		IsContract:        addr.IsContract,
		TxCount:           addr.TxsCount,
		DeployedContracts: addr.ContractsCount,
		Interactions:      addr.Interactions,
		Balance: Balance{
			Value: addr.Balance.Value.String(),
		},
	}
	return address
}

// Balance info
//
//	@Description	Balance of address information
type Balance struct {
	Value string `example:"10000000000" json:"value" swaggertype:"string"`
}

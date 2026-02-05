package responses

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
)

// BeaconWithdrawal represents a beacon chain (consensus layer) withdrawal
// @Description Beacon chain withdrawal transferring ETH from a validator to an execution layer address
type BeaconWithdrawal struct {
	Height         uint64    `example:"100"                                       json:"height"          swaggertype:"integer"`
	Time           time.Time `example:"2023-07-04T03:10:57+00:00"                 json:"time"            swaggertype:"string"`
	Index          int64     `example:"1"                                         json:"index"           swaggertype:"integer"`
	ValidatorIndex int64     `example:"100000"                                    json:"validator_index" swaggertype:"integer"`
	Address        string    `example:"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb" json:"address"         swaggertype:"string"`
	Amount         string    `example:"1000000000000000000"                       json:"amount"          swaggertype:"string"`
}

func NewBeaconWithdrawal(bw *storage.BeaconWithdrawal) BeaconWithdrawal {
	return BeaconWithdrawal{
		Height:         uint64(bw.Height),
		Time:           bw.Time,
		Index:          bw.Index,
		ValidatorIndex: bw.ValidatorIndex,
		Address:        bw.Address.Hash.String(),
		Amount:         bw.Amount.String(),
	}
}

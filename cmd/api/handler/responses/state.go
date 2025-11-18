package responses

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
)

type State struct {
	Id            uint64         `example:"321"                                                                format:"int64"     json:"id"             swaggertype:"integer"`
	Name          string         `example:"indexer"                                                            format:"string"    json:"name"           swaggertype:"string"`
	LastHeight    pkgTypes.Level `example:"100"                                                                format:"int64"     json:"last_height"    swaggertype:"integer"`
	LastHash      string         `example:"0x85480d3bbf5d757b63375ab9da566e7c330e2b6b9abe965fc7f41542d3edaeaa" format:"string"    json:"hash"           swaggertype:"string"`
	LastTime      time.Time      `example:"2023-07-04T03:10:57+00:00"                                          format:"date-time" json:"last_time"      swaggertype:"string"`
	TotalTx       int64          `example:"23456"                                                              format:"int64"     json:"total_tx"       swaggertype:"integer"`
	TotalAccounts int64          `example:"43"                                                                 format:"int64"     json:"total_accounts" swaggertype:"integer"`
}

func NewState(state storage.State) State {
	return State{
		Id:            state.Id,
		Name:          state.Name,
		LastHeight:    state.LastHeight,
		LastHash:      pkgTypes.Hex(state.LastHash).Hex(),
		LastTime:      state.LastTime,
		TotalTx:       state.TotalTx,
		TotalAccounts: state.TotalAccounts,
	}
}

package responses

import (
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
)

// Log model info
//
//	@Description	Token transfer information
type Log struct {
	Id      uint64      `example:"0"                                                                  json:"id"            swaggertype:"integer"`
	Address string      `example:"0x0000000000000000000000000000000000000001"                         json:"address"       swaggertype:"string"`
	Height  uint64      `example:"100"                                                                json:"height"        swaggertype:"integer"`
	Time    time.Time   `example:"2026-01-01T01:01:01+00:00"                                          format:"date-time"   json:"time"           swaggertype:"string"`
	TxHash  string      `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"tx_hash"       swaggertype:"string"`
	Data    string      `example:"0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF" json:"data"          swaggertype:"string"`
	Name    string      `example:"0xff64905f73a67fb594e0f940a8075a860db489ad991e032f48c81123eb52d60b" json:"name"          swaggertype:"string"`
	Index   int64       `example:"1"                                                                  json:"index"         swaggertype:"integer"`
	Topics  []string    `json:"topics"`
	Decoded *DecodedLog `json:"decoded,omitempty"                                                     swaggertype:"object"`
}

func NewLog(log storage.Log) Log {
	l := Log{
		Id:      log.Id,
		Address: log.Address.Hash.Hex(),
		Height:  uint64(log.Height),
		Time:    log.Time,
		TxHash:  log.Tx.Hash.Hex(),
		Data:    log.Data.Hex(),
		Name:    log.Name,
		Index:   log.Index,
	}

	for _, topic := range log.Topics {
		l.Topics = append(l.Topics, topic.Hex())
	}

	if parsedABI := parseABI(log.ContractABI); parsedABI != nil {
		l.Decoded = decodeLogWithABI(parsedABI, log.Data, log.Topics)
	}

	return l
}

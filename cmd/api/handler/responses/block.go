package responses

import (
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type Block struct {
	Height               uint64          `example:"100"                                                                json:"height"                 swaggertype:"integer"`
	Time                 time.Time       `example:"2023-07-04T03:10:57+00:00"                                          json:"time"                   swaggertype:"string"`
	GasLimit             decimal.Decimal `example:"1000000"                                                            json:"gas_limit"              swaggertype:"integer"`
	GasUsed              decimal.Decimal `example:"500000"                                                             json:"gas_used"               swaggertype:"integer"`
	Hash                 string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"hash"                   swaggertype:"string"`
	ParentHash           string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"parent_hash"            swaggertype:"string"`
	Difficulty           string          `example:"0x0"                                                                json:"difficulty"             swaggertype:"string"`
	ExtraData            string          `example:"0x726574682f76312e372e302f6c696e7578"                               json:"extra_data"             swaggertype:"string"`
	LogsBloom            string          `example:"0x0000000000000000000020000000000"                                  json:"logs_bloom"             swaggertype:"string"`
	Miner                string          `example:"0x0000000000000000000000000000000000000000"                         json:"miner"                  swaggertype:"string"`
	MixHash              string          `example:"0x000000000000000000000000000000000000000000000000000000000033a87e" json:"mix_hash"               swaggertype:"string"`
	Nonce                uint64          `example:"0"                                                                  json:"nonce"                  swaggertype:"integer"`
	ReceiptsRoot         string          `example:"0x24e9aae3033f9ff809675831eca331b701440009592a40a6d788756f3be983a2" json:"receipts_root"          swaggertype:"string"`
	Sha3Uncles           string          `example:"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347" json:"sha3_uncles_hash"       swaggertype:"string"`
	Size                 uint64          `example:"0"                                                                  json:"size"                   swaggertype:"integer"`
	StateRoot            string          `example:"0x9b6e76e8263c5060b61e396c65baf15dd187386d5607250be0dcc5308f0b49ef" json:"state_root"             swaggertype:"string"`
	TransactionsRootHash string          `example:"0x0764012270afacd3b101bcfadaaa9fc3190d04ed90ff22c0ee59781e54858a7d" json:"transactions_root_hash" swaggertype:"string"`
	Stats                *BlockStats     `json:"stats,omitempty"`
}

func NewBlock(block storage.Block) Block {
	resultBlock := Block{
		Height:               uint64(block.Height),
		Time:                 block.Time,
		GasLimit:             block.GasLimit,
		GasUsed:              block.GasUsed,
		Hash:                 block.Hash.Hex(),
		ParentHash:           block.ParentHashHash.Hex(),
		Difficulty:           block.DifficultyHash.Hex(),
		ExtraData:            block.ExtraDataHash.Hex(),
		LogsBloom:            block.LogsBloomHash.Hex(),
		Miner:                block.Miner.Hash.Hex(),
		MixHash:              block.MixHash.Hex(),
		ReceiptsRoot:         block.ReceiptsRootHash.Hex(),
		Sha3Uncles:           block.Sha3UnclesHash.Hex(),
		StateRoot:            block.StateRootHash.Hex(),
		TransactionsRootHash: block.TransactionsRootHash.Hex(),
	}

	size, err := block.SizeHash.Uint64()
	if err != nil {
		log.Err(err).Uint64("block_height", uint64(block.Height)).Msg("converting block size")
	}
	resultBlock.Size = size

	nonce, err := block.NonceHash.Uint64()
	if err != nil {
		log.Err(err).Uint64("block_height", uint64(block.Height)).Msg("converting block nonce")
	}
	resultBlock.Nonce = nonce

	if block.Stats != nil {
		resultBlock.Stats = NewBlockStats(*block.Stats)
	}
	return resultBlock
}

type BlockStats struct {
	Height    uint64    `example:"100"                       json:"height"     swaggertype:"integer"`
	Time      time.Time `example:"2023-07-04T03:10:57+00:00" json:"time"       swaggertype:"string"`
	TxCount   int64     `example:"12"                        json:"tx_count"   swaggertype:"integer"`
	BlockTime uint64    `example:"1000"                      json:"block_time" swaggertype:"integer"`
}

func NewBlockStats(stats storage.BlockStats) *BlockStats {
	return &BlockStats{
		Height:    uint64(stats.Height),
		Time:      stats.Time,
		TxCount:   stats.TxCount,
		BlockTime: stats.BlockTime,
	}
}

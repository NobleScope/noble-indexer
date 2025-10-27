package types

type Block struct {
	Difficulty       Hex   `json:"difficulty"`
	ExtraData        Hex   `json:"extraData"`
	GasLimit         Hex   `json:"gasLimit"`
	GasUsed          Hex   `json:"gasUsed"`
	BaseFeePerGas    Hex   `json:"baseFeePerGas"`
	Hash             Hex   `json:"hash"`
	LogsBloom        Hex   `json:"logsBloom"`
	Miner            Hex   `json:"miner"`
	MixHash          Hex   `json:"mixHash"`
	Nonce            Hex   `json:"nonce"`
	Number           Hex   `json:"number"`
	ParentHash       Hex   `json:"parentHash"`
	ReceiptsRoot     Hex   `json:"receiptsRoot"`
	Sha3Uncles       Hex   `json:"sha3Uncles"`
	Size             Hex   `json:"size"`
	StateRoot        Hex   `json:"stateRoot"`
	Timestamp        Hex   `json:"timestamp"`
	TotalDifficulty  Hex   `json:"totalDifficulty"`
	Transactions     []Tx  `json:"transactions"`
	TransactionsRoot Hex   `json:"transactionsRoot"`
	Uncles           []Hex `json:"uncles"`
}

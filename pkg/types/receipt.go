package types

type Receipt struct {
	BlockHash         Hex   `json:"blockHash"`
	BlockNumber       Hex   `json:"blockNumber"`
	ContractAddress   *Hex  `json:"contractAddress"`
	CumulativeGasUsed Hex   `json:"cumulativeGasUsed"`
	EffectiveGasPrice Hex   `json:"effectiveGasPrice"`
	From              Hex   `json:"from"`
	GasUsed           Hex   `json:"gasUsed"`
	L1Fee             Hex   `json:"l1Fee"`
	Logs              []Log `json:"logs"`
	LogsBloom         Hex   `json:"logsBloom"`
	Status            Hex   `json:"status"`
	To                Hex   `json:"to"`
	TransactionHash   Hex   `json:"transactionHash"`
	TransactionIndex  Hex   `json:"transactionIndex"`
	Type              Hex   `json:"type"`
}

type Log struct {
	Address          Hex   `json:"address"`
	Topics           []Hex `json:"topics"`
	Data             Hex   `json:"data"`
	BlockNumber      Hex   `json:"blockNumber"`
	TransactionHash  Hex   `json:"transactionHash"`
	TransactionIndex Hex   `json:"transactionIndex"`
	BlockHash        Hex   `json:"blockHash"`
	LogIndex         Hex   `json:"logIndex"`
	Removed          bool  `json:"removed"`
}

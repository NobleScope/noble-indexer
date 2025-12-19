package types

type Trace struct {
	Action       Action      `json:"action"`
	BlockHash    Hex         `json:"blockHash"`
	BlockNumber  uint64      `json:"blockNumber"`
	Result       TraceResult `json:"result"`
	Subtraces    uint64      `json:"subtraces"`
	TraceAddress []uint64    `json:"traceAddress"`
	TxHash       *Hex        `json:"transactionHash"`
	TxPosition   *uint64     `json:"transactionPosition"`
	Type         string      `json:"type"`
}

type Action struct {
	From           *Hex    `json:"from"`
	Gas            *Hex    `json:"gas"`
	To             *Hex    `json:"to"`
	Init           *Hex    `json:"init"`
	Value          *Hex    `json:"value"`
	CreationMethod *string `json:"creationMethod"`
	CallType       *string `json:"callType"`
	Input          *Hex    `json:"input"`
	Author         *Hex    `json:"author"`
	RewardType     *string `json:"rewardType"`
}

type TraceResult struct {
	GasUsed Hex  `json:"gasUsed"`
	Address *Hex `json:"address"`
	Code    *Hex `json:"code"`
	Output  *Hex `json:"output"`
}

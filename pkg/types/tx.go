package types

type Tx struct {
	BlockHash        Hex `json:"blockHash"`
	BlockNumber      Hex `json:"blockNumber"`
	From             Hex `json:"from"`
	Gas              Hex `json:"gas"`
	GasPrice         Hex `json:"gasPrice"`
	Hash             Hex `json:"hash"`
	Input            Hex `json:"input"`
	Nonce            Hex `json:"nonce"`
	To               Hex `json:"to"`
	TransactionIndex Hex `json:"transactionIndex"`
	Value            Hex `json:"value"`
	Type             Hex `json:"type"`
	V                Hex `json:"v"`
	R                Hex `json:"r"`
	S                Hex `json:"s"`
}

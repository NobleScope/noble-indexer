package responses

type SearchItem struct {
	// Result type which is in the result. Can be 'address', 'block', 'tx', 'token'
	Type string `json:"type"`

	// Search result. Can be one of folowwing types: Address, Block, Tx, Token
	Result any `json:"result" swaggertype:"object"`
}

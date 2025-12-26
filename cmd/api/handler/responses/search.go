package responses

// SearchItem model info
//
//	@Description	Search result item
type SearchItem struct {
	// Result type which is in the result. Can be 'address', 'block', 'tx', 'token'
	Type string `enums:"address,block,tx,token" example:"address" json:"type" swaggertype:"string"`

	// Search result. Can be one of folowwing types: Address, Block, Tx, Token
	Result any `json:"result" swaggertype:"object"`
}

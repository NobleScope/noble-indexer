package types

// swagger:enum TokenType
/*
	ENUM(
		ERC20
		ERC721
		ERC1155
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type TokenType string

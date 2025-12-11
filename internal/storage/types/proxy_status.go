package types

// swagger:enum ProxyStatus
/*
	ENUM(
		new
		resolved
		error
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type ProxyStatus string

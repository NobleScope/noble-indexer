package types

// swagger:enum CallType
/*
	ENUM(
		call,
		delegatecall,
		staticcall,
		callcode
	)
*/
//go:generate go-enum --marshal --sql --values --names
type CallType string

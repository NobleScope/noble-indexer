package types

// swagger:enum TraceType
/*
	ENUM(
		call,
		create,
		create2,
		selfdestruct,
		reward,
		suicide
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type TraceType string

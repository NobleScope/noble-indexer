package types

// swagger:enum TransferType
/*
	ENUM(
		burn
		mint
		transfer
		unknown
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type TransferType string

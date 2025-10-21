package types

// swagger:enum TxStatus
/*
	ENUM(
		TxStatusSuccess
		TxStatusRevert
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type TxStatus string

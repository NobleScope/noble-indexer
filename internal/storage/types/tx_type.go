package types

// swagger:enum TxType
/*
	ENUM(
		TxTypeUnknown,
		TxTypeLegacy,
		TxTypeDynamicFee
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type TxType string

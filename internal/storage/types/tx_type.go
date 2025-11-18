package types

// swagger:enum TxType
/*
	ENUM(
		TxTypeUnknown,
		TxTypeLegacy,
		TxTypeDynamicFee,
		TxTypeBlob,
		TxTypeSetCode
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type TxType string

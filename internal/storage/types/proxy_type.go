package types

// swagger:enum ProxyType
/*
	ENUM(
		EIP1167
		EIP7760
		EIP7702
		EIP1967
		custom
		clone_with_immutable_args
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type ProxyType string

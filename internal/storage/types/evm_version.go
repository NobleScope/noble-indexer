package types

// swagger:enum EVMVersion
/*
	ENUM(
		homestead
		tangerineWhistle
		spuriousDragon
		byzantium
		constantinople
		petersburg
		istanbul
		berlin
		london
		paris
		shanghai
		cancun
		prague
		osaka
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type EVMVersion string

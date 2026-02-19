package types

// swagger:enum LicenseType
/*
	ENUM(
		none
		unlicense
		mit
		gnu_gpl_v2
		gnu_gpl_v3
		gnu_lgpl_v2_1
		gnu_lgpl_v3
		bsd_2_clause
		bsd_3_clause
		mpl_2_0
		osl_3_0
		apache_2_0
		gnu_agpl_v3
		bsl_1_1
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type LicenseType string

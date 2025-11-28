package types

// swagger:enum MetadataStatus
/*
	ENUM(
		success
		failed
		pending
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type MetadataStatus string

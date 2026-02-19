package types

// swagger:enum VerificationTaskStatus
/*
	ENUM(
		VerificationStatusNew
		VerificationStatusPending
		VerificationStatusFailed
		VerificationStatusSuccess
	)
*/
//go:generate go-enum --marshal --sql --values --noprefix --names
type VerificationTaskStatus string

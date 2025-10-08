package types

type Payload struct {
	Subscription Hex     `json:"subscription"`
	Result       *Result `json:"result"`
}
type Result struct {
	Number Hex `json:"number"`
}

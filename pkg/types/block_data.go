package types

type BlockData struct {
	Block
	Receipts []Receipt
	Traces   []Trace
}

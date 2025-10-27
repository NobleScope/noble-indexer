package enum

type BalanceOp string

const (
	Add BalanceOp = "add"
	Sub BalanceOp = "sub"
)

func (op BalanceOp) String() string {
	switch op {
	case Add:
		return "add"
	case Sub:
		return "sub"
	default:
		return "unknown"
	}
}

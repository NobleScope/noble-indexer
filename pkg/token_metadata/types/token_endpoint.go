package types

type TokenEndpoint int

const (
	Name TokenEndpoint = iota
	Symbol
	Decimals
	Uri
	TokenUri
)

func (t TokenEndpoint) String() string {
	switch t {
	case Name:
		return "name"
	case Symbol:
		return "symbol"
	case Decimals:
		return "decimals"
	case Uri:
		return "uri"
	case TokenUri:
		return "tokenURI"
	default:
		return "unknown"
	}
}

package types

import "regexp"

var (
	ContractNameRe    = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	CompilerVersionRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+(\+commit\.[0-9a-f]+)?$`)
)

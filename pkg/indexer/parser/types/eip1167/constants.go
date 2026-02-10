package eip1167

import pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"

const CodeLength = 45

var (
	Prefix  = pkgTypes.MustDecodeHex("0x363d3d373d3d3d363d73")
	Postfix = pkgTypes.MustDecodeHex("0x5af43d82803e903d91602b57fd5bf3")
)

package eip7760_basic20

import pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"

var (
	Prefix  = pkgTypes.MustDecodeHex("0x3d3d3373")
	Postfix = pkgTypes.MustDecodeHex("0x14605757363d3d37363d7f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3" +
		"ca505d382bbc545af43d6000803e6052573d6000fd5b3d6000f35b3d356020355560408036111560525736038060403d373d3d355af4" +
		"3d6000803e6052573d6000fd")
	Length = len(Prefix) + 20 + len(Postfix)
)

package eip7760_basic14

import (
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
)

var (
	Prefix  = pkgTypes.MustDecodeHex("0x3d3d336d")
	Postfix = pkgTypes.MustDecodeHex("0x14605157363d3d37363d7f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3" +
		"ca505d382bbc545af43d6000803e604c573d6000fd5b3d6000f35b3d3560203555604080361115604c5736038060403d373d3d355af4" +
		"3d6000803e604c573d6000fd")
	Length = len(Prefix) + 14 + len(Postfix)
)

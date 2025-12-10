package eip7760_I_14

import pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"

var (
	Prefix  = pkgTypes.MustDecodeHex("0x365814607d573d3d336d")
	Postfix = pkgTypes.MustDecodeHex("0x14605757363d3d37363d7f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3" +
		"ca505d382bbc545af43d6000803e6052573d6000fd5b3d6000f35b3d35602035556040360380156052578060403d373d3d355af43d60" +
		"00803e6052573d6000fd5b602060233d393d51543d52593df3")
	Length = len(Prefix) + 14 + len(Postfix)
)

package eip7760_I_20

import pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"

var (
	Prefix  = pkgTypes.MustDecodeHex("0x3658146083573d3d3373")
	Postfix = pkgTypes.MustDecodeHex("0x14605d57363d3d37363d7f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3" +
		"ca505d382bbc545af43d6000803e6058573d6000fd5b3d6000f35b3d35602035556040360380156058578060403d373d3d355af43d60" +
		"00803e6058573d6000fd5b602060293d393d51543d52593df3")
	Length = len(Prefix) + 20 + len(Postfix)
)

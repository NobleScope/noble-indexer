package clone_with_immutable_args

import pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"

var (
	First = pkgTypes.MustDecodeHex("0x3d3d3d3d363d3d3761")
	Third = pkgTypes.MustDecodeHex("0x603736393661")
	Fifth = pkgTypes.MustDecodeHex("0x013d73")

	FirstLen = len(First)
	ThirdLen = len(Third)
	FifthLen = len(Fifth)

	ThirdStart = FirstLen + 2
	ThirdEnd   = ThirdStart + ThirdLen
	FifthStart = ThirdEnd + 2
	FifthEnd   = FifthStart + FifthLen

	MinimalCodeLength = 42
)

package custom_v1_0_0

import pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"

var (
	Code = pkgTypes.MustDecodeHex("0x608060405273ffffffffffffffffffffffffffffffffffffffff60005416366000803760008" +
		"0366000845af43d6000803e6000811415603d573d6000fd5b3d6000f3fea165627a7a723058201e7d648b83cfac072cbccefc2ffc62a" +
		"6999d4a050ee87a721942de1da9670db80029")
	Length = len(Code)
)

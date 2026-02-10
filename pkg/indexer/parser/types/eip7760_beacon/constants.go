package eip7760_beacon

import pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"

var (
	Code = pkgTypes.MustDecodeHex("0x363d3d373d3d363d602036600436635c60da1b60e01b36527fa3f0ad74e5423aebfd80d3ef434" +
		"6578335a9a72aeaee59ff6cb3582b35133d50545afa5036515af43d6000803e604d573d6000fd5b3d6000f3")
	Length = len(Code)
)

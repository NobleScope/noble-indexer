package types

import (
	"math/big"

	internalTypes "github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type TokenMetadata struct {
	Name     Hex
	Symbol   Hex
	Decimals Hex
	URI      Hex
}

type TokenMetadataRequest struct {
	Id        uint64
	Address   string
	ABI       abi.ABI
	Interface internalTypes.TokenType
	TokenID   *big.Int
}

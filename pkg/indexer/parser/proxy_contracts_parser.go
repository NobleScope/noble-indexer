package parser

import (
	"bytes"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	clone "github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/clone_with_immutable_args"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/custom_v1_0_0"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/custom_v1_1_1"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/custom_v1_3_0"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/custom_v1_3_0_zksync"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/custom_v1_4_1"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/custom_v1_4_1_zksync"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/custom_v1_5_0"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip1167"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip1967"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_I_14"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_I_20"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_basic14"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_basic20"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_beacon"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_beacon_I"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_uups"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/eip7760_uups_I"
)

func (p *Module) parseProxyContract(ctx *dCtx.Context, contract *storage.Contract) error {
	proxyType := getProxyType(contract)

	switch proxyType {
	case types.EIP1167:
		implementationAddress := storage.Address{
			Address:    contract.Code[len(eip1167.Prefix) : len(eip1167.Prefix)+AddressBytesLength].String(),
			LastHeight: ctx.Block.Height,
			IsContract: true,
			Balance:    storage.EmptyBalance(),
		}
		ctx.AddProxyContract(&storage.ProxyContract{
			Height: ctx.Block.Height,
			Type:   proxyType,
			Status: types.Resolved,
			Contract: storage.Contract{
				Address: contract.Address,
			},
			Implementation: &storage.Contract{
				Address: implementationAddress,
			},
		})
		ctx.AddAddress(&implementationAddress)
	case types.EIP7760:
		ctx.AddProxyContract(&storage.ProxyContract{
			Height: ctx.Block.Height,
			Type:   proxyType,
			Status: types.New,
			Contract: storage.Contract{
				Address: contract.Address,
			},
		})
	case types.Custom:
		ctx.AddProxyContract(&storage.ProxyContract{
			Height: ctx.Block.Height,
			Type:   proxyType,
			Status: types.New,
			Contract: storage.Contract{
				Address: contract.Address,
			},
		})
	case types.CloneWithImmutableArgs:
		implementationAddress := storage.Address{
			Address:    contract.Code[clone.FifthEnd : clone.FifthEnd+AddressBytesLength].String(),
			LastHeight: ctx.Block.Height,
			IsContract: true,
			Balance:    storage.EmptyBalance(),
		}
		ctx.AddProxyContract(&storage.ProxyContract{
			Height: ctx.Block.Height,
			Type:   proxyType,
			Status: types.Resolved,
			Contract: storage.Contract{
				Address: contract.Address,
			},
			Implementation: &storage.Contract{
				Address: implementationAddress,
			},
		})
		ctx.AddAddress(&implementationAddress)
	}
	return nil
}

func (p *Module) parseEIP1967Proxy(ctx *dCtx.Context, logs []*storage.Log) {
	eip1967Contracts := getEIP1967Proxy(logs)
	for proxyAddress, implementationAddress := range eip1967Contracts {
		storageImplementationAddress := storage.Address{
			Address:    implementationAddress,
			LastHeight: ctx.Block.Height,
			IsContract: true,
			Balance:    storage.EmptyBalance(),
		}

		ctx.AddProxyContract(&storage.ProxyContract{
			Height: ctx.Block.Height,
			Type:   types.EIP1967,
			Status: types.Resolved,
			Contract: storage.Contract{
				Address: storage.Address{
					Address: proxyAddress,
				},
			},
			Implementation: &storage.Contract{
				Address: storageImplementationAddress,
			},
		})

		ctx.AddAddress(&storageImplementationAddress)
	}
}

func getProxyType(contract *storage.Contract) types.ProxyType {
	if isEIP1167(contract) {
		return types.EIP1167
	}
	if isEIP7760(contract) {
		return types.EIP7760
	}
	if isCustom(contract) {
		return types.Custom
	}
	if isCloneWithImmutableArgs(contract) {
		return types.CloneWithImmutableArgs
	}

	return ""
}

func isEIP1167(contract *storage.Contract) bool {
	if len(contract.Code) != eip1167.CodeLength {
		return false
	}

	return bytes.Equal(contract.Code[:len(eip1167.Prefix)], eip1167.Prefix) &&
		bytes.Equal(contract.Code[len(eip1167.Prefix)+AddressBytesLength:], eip1167.Postfix)
}

func isEIP7760(contract *storage.Contract) bool {
	if isBasic20 := len(contract.Code) == eip7760_basic20.Length &&
		bytes.Equal(contract.Code[:len(eip7760_basic20.Prefix)], eip7760_basic20.Prefix) &&
		bytes.Equal(contract.Code[:len(eip7760_basic20.Postfix)], eip7760_basic20.Postfix); isBasic20 {
		return true
	}

	if isBasic14 := len(contract.Code) == eip7760_basic14.Length &&
		bytes.Equal(contract.Code[:len(eip7760_basic14.Prefix)], eip7760_basic14.Prefix) &&
		bytes.Equal(contract.Code[:len(eip7760_basic14.Postfix)], eip7760_basic14.Postfix); isBasic14 {
		return true
	}

	if isIVariant20 := len(contract.Code) == eip7760_I_20.Length &&
		bytes.Equal(contract.Code[:len(eip7760_I_20.Prefix)], eip7760_I_20.Prefix) &&
		bytes.Equal(contract.Code[:len(eip7760_I_20.Postfix)], eip7760_I_20.Postfix); isIVariant20 {
		return true
	}

	if isIVariant14 := len(contract.Code) == eip7760_I_14.Length &&
		bytes.Equal(contract.Code[:len(eip7760_I_14.Prefix)], eip7760_I_14.Prefix) &&
		bytes.Equal(contract.Code[:len(eip7760_I_14.Postfix)], eip7760_I_14.Postfix); isIVariant14 {
		return true
	}

	if isUUPS := len(contract.Code) == eip7760_uups.Length && bytes.Equal(contract.Code, eip7760_uups.Code); isUUPS {
		return true
	}

	if isUUPSIVariant := len(contract.Code) == eip7760_uups_I.Length &&
		bytes.Equal(contract.Code, eip7760_uups_I.Code); isUUPSIVariant {
		return true
	}

	if isBeacon := len(contract.Code) == eip7760_beacon.Length &&
		bytes.Equal(contract.Code, eip7760_beacon.Code); isBeacon {
		return true
	}

	if isBeaconIVariant := len(contract.Code) == eip7760_beacon_I.Length &&
		bytes.Equal(contract.Code, eip7760_beacon_I.Code); isBeaconIVariant {
		return true
	}

	return false
}

func isCustom(contract *storage.Contract) bool {
	if isCustomV1_0_0 := len(contract.Code) == custom_v1_0_0.Length &&
		bytes.Equal(contract.Code, custom_v1_0_0.Code); isCustomV1_0_0 {
		return true
	}

	if isCustomV1_1_1 := len(contract.Code) == custom_v1_1_1.Length &&
		bytes.Equal(contract.Code, custom_v1_1_1.Code); isCustomV1_1_1 {
		return true
	}

	if isCustomV1_3_0 := len(contract.Code) == custom_v1_3_0.Length &&
		bytes.Equal(contract.Code, custom_v1_3_0.Code); isCustomV1_3_0 {
		return true
	}

	if isCustomV1_3_0_zksync := len(contract.Code) == custom_v1_3_0_zksync.Length &&
		bytes.Equal(contract.Code, custom_v1_3_0_zksync.Code); isCustomV1_3_0_zksync {
		return true
	}

	if isCustomV1_4_1 := len(contract.Code) == custom_v1_4_1.Length &&
		bytes.Equal(contract.Code, custom_v1_4_1.Code); isCustomV1_4_1 {
		return true
	}

	if isCustomV1_4_1_zksync := len(contract.Code) == custom_v1_4_1_zksync.Length &&
		bytes.Equal(contract.Code, custom_v1_4_1_zksync.Code); isCustomV1_4_1_zksync {
		return true
	}

	if isCustomV1_5_0 := len(contract.Code) == custom_v1_5_0.Length &&
		bytes.Equal(contract.Code, custom_v1_5_0.Code); isCustomV1_5_0 {
		return true
	}

	return false
}

func isCloneWithImmutableArgs(contract *storage.Contract) bool {
	return len(contract.Code) >= clone.MinimalCodeLength &&
		bytes.Equal(contract.Code[:clone.FirstLen], clone.First) &&
		bytes.Equal(contract.Code[clone.ThirdStart:clone.ThirdEnd], clone.Third) &&
		bytes.Equal(contract.Code[clone.FifthStart:clone.FifthEnd], clone.Fifth)
}

func getEIP1967Proxy(logs []*storage.Log) map[string]string {
	result := make(map[string]string)
	if len(logs) == 0 {
		return result
	}

	for i := range logs {
		if len(logs[i].Topics) == 0 {
			continue
		}

		for range logs[i].Topics {
			if len(logs[i].Topics) > 1 && (logs[i].Topics[0].Hex() == eip1967.EventUpgradedSignature ||
				logs[i].Topics[0].Hex() == eip1967.EventBeaconUpgradedSignature) {
				implementationAddress := logs[i].Topics[1]
				if len(implementationAddress) > AddressBytesLength {
					implementationAddress = implementationAddress[len(implementationAddress)-AddressBytesLength:]
				}
				result[logs[i].Address.String()] = implementationAddress.String()
			}
		}
	}

	return result
}

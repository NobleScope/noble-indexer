package proxy_contracts_resolver

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/pkg/errors"
)

func (p *Module) resolveProxyContracts(ctx context.Context, contracts []*storage.ProxyContract) (int, error) {
	storageRequest := make([]pkgTypes.StorageRequest, len(contracts))
	resolved := 0

	for i := range contracts {
		hexAddress, err := pkgTypes.HexFromString(contracts[i].Contract.Address.Address)
		if err != nil {
			return resolved, errors.Wrapf(err, "conversing proxy contract address")
		}
		storageRequest[i].BlockNumber = contracts[i].Height
		storageRequest[i].StorageSlot = getStorageSlot(contracts[i].Type)
		storageRequest[i].ContractAddress = hexAddress.Hex()
	}

	storageValues, err := p.api.Storage(ctx, storageRequest)
	if err != nil {
		p.handleError(contracts)
		return resolved, errors.Wrap(err, "fetching contracts storage")
	}
	if len(storageValues) != len(contracts) {
		p.handleError(contracts)
		return resolved, errors.New("unexpected number of storage values")
	}

	for i := range contracts {
		contracts[i].ResolvingAttempts += 1
		if contracts[i].ResolvingAttempts >= p.cfg.Proxy.MaxResolvingAttempts {
			contracts[i].Status = types.Error
			continue
		}

		implementationAddress := storageValues[i]
		if implementationAddress == nil {
			continue
		}

		if len(implementationAddress) > parser.AddressBytesLength {
			implementationAddress = implementationAddress[len(implementationAddress)-parser.AddressBytesLength:]
		}
		contracts[i].Status = types.Resolved
		contracts[i].Implementation = &storage.Contract{
			Address: storage.Address{Address: implementationAddress.String()},
		}
		resolved += 1
	}

	p.MustOutput(OutputName).Push(contracts)
	return resolved, nil
}

func getStorageSlot(proxyType types.ProxyType) string {
	switch proxyType {
	case types.EIP7760:
		return "0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"
	case types.Custom:
		return "0x0"
	}
	return ""
}

func (p *Module) handleError(contracts []*storage.ProxyContract) {
	for i := range contracts {
		contracts[i].ResolvingAttempts += 1
	}
	p.MustOutput(OutputName).Push(contracts)
}

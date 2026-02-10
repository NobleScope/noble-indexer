package storage

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/postgres"
	"github.com/pkg/errors"
)

func (module *Module) listenProxyImplementations(ctx context.Context) {
	module.Log.Info().Msg("listening proxy implementations")
	input := module.MustInput(ProxyContractsInput)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-input.Listen():
			if !ok {
				module.Log.Warn().Msg("can't read message from input")
				module.MustOutput(StopOutput).Push(struct{}{})
				continue
			}

			proxyContracts, ok := msg.([]*storage.ProxyContract)
			if !ok {
				module.Log.Warn().Msgf("invalid message type: %T", msg)
				continue
			}

			err := module.updateProxyContracts(ctx, proxyContracts)
			if err != nil {
				module.Log.Error().Err(err).Msg("updating proxy contracts implementations")
				module.MustOutput(StopOutput).Push(struct{}{})
				continue
			}
		}
	}
}

func saveProxyContracts(
	ctx context.Context,
	tx storage.Transaction,
	proxyContracts []*storage.ProxyContract,
	addresses map[string]uint64,
) error {
	if len(proxyContracts) == 0 {
		return nil
	}

	for i := range proxyContracts {
		contractID, ok := addresses[proxyContracts[i].Contract.Address.String()]
		if !ok {
			return errors.Errorf("can't find contract key: %s", proxyContracts[i].Contract.Address.String())
		}
		proxyContracts[i].Id = contractID

		if proxyContracts[i].Implementation == nil {
			continue
		}

		implementationID, ok := addresses[proxyContracts[i].Implementation.Address.String()]
		if !ok {
			return errors.Errorf("can't find contract implementation key: %s", proxyContracts[i].Implementation.Address.String())
		}
		proxyContracts[i].ImplementationID = &implementationID
	}

	if err := tx.SaveProxyContracts(ctx, proxyContracts...); err != nil {
		return errors.Wrap(err, "saving proxy contracts")
	}

	return nil
}

func (module *Module) updateProxyContracts(ctx context.Context, proxyContracts []*storage.ProxyContract) error {
	tx, err := postgres.BeginTransaction(ctx, module.storage)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	addressesMap := make(map[string]*storage.Address, len(proxyContracts))
	for i := range proxyContracts {
		proxyAddress := proxyContracts[i].Contract.Address
		if _, ok := addressesMap[proxyAddress.String()]; !ok {
			addressesMap[proxyAddress.String()] = &storage.Address{
				Hash:       proxyAddress.Hash,
				IsContract: true,
				Balance:    storage.EmptyBalance(),
			}
		}
		if proxyContracts[i].Implementation == nil {
			continue
		}
		implementationAddress := proxyContracts[i].Implementation.Address
		if _, ok := addressesMap[implementationAddress.String()]; !ok {
			addressesMap[implementationAddress.String()] = &storage.Address{
				Hash:       implementationAddress.Hash,
				IsContract: true,
				Balance:    storage.EmptyBalance(),
			}
		}
	}

	addresses := make([]*storage.Address, 0, len(addressesMap))
	for _, address := range addressesMap {
		addresses = append(addresses, address)
	}

	addrToId, _, err := saveAddresses(ctx, tx, addresses)
	if err != nil {
		return errors.Wrap(err, "saving proxy contracts addresses")
	}

	if err = saveProxyContracts(ctx, tx, proxyContracts, addrToId); err != nil {
		return err
	}
	if err = tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	return nil
}

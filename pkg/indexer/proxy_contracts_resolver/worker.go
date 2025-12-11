package proxy_contracts_resolver

import (
	"context"
	"slices"

	"github.com/baking-bad/noble-indexer/internal/storage"
)

func (p *Module) dispatcher(ctx context.Context) ([][]*storage.ProxyContract, error) {
	contracts, err := p.pg.ProxyContracts.NotResolved(ctx)
	if err != nil {
		return nil, err
	}
	if len(contracts) == 0 {
		return nil, nil
	}

	filtered := make([]*storage.ProxyContract, 0, len(contracts))
	for i := range contracts {
		if _, exists := p.taskQueue.Get(contracts[i].String()); !exists {
			p.taskQueue.Set(contracts[i].String(), struct{}{})
			filtered = append(filtered, &contracts[i])
		}
	}

	batches := slices.Collect(slices.Chunk(filtered, p.cfg.Proxy.BatchSize))
	p.Log.Info().
		Int("total", len(contracts)).
		Int("new", len(filtered)).
		Int("batches", len(batches)).
		Msg("fetched unresolved proxy contracts")

	return batches, nil
}

func (p *Module) worker(ctx context.Context, contracts []*storage.ProxyContract) {
	defer p.clearQueue(contracts)

	p.Log.Debug().
		Int("batch_size", len(contracts)).
		Msg("processing batch")

	resolved, err := p.resolveProxyContracts(ctx, contracts)
	if err != nil {
		p.Log.Error().
			Err(err).
			Int("batch_size", len(contracts)).
			Msg("failed to resolve proxy contracts batch")
		return
	}

	p.Log.Info().
		Int("batch_size", len(contracts)).
		Int("resolved", resolved).
		Msg("proxy contracts batch resolved")
}

func (p *Module) clearQueue(contracts []*storage.ProxyContract) {
	for i := range contracts {
		p.taskQueue.Delete(contracts[i].String())
	}
}

func (p *Module) dispatcherErrorHandler(_ context.Context, err error) {
	p.Log.Error().
		Err(err).
		Msg("dispatching proxy contracts")
}

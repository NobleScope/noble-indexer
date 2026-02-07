package contract_metadata

import (
	"context"
	"path"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/baking-bad/noble-indexer/internal/cache"
	"github.com/baking-bad/noble-indexer/internal/ipfs"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/pkg/errors"
)

type Module struct {
	modules.BaseModule

	pool        ipfs.Pool
	pg          postgres.Storage
	storage     sdk.Transactable
	syncPeriod  time.Duration
	retryDelay  time.Duration
	cfg         config.Config
	abiRegistry *ABIRegistry
}

func NewModule(pg postgres.Storage, cfg config.Config) *Module {
	opts := make([]ipfs.Option, 0)
	if cfg.Cache.URL != "" {
		if cfg.Cache.TTL <= 0 {
			cfg.Cache.TTL = 1800 // 30 minutes
		}
		cache, err := cache.NewValKey(cfg.Cache.URL, time.Duration(cfg.Cache.TTL)*time.Second)
		if err != nil {
			panic(err)
		}
		opts = append(opts, ipfs.WithCache(cache))
	}
	pool, err := ipfs.New(cfg.ContractMetadataResolver.MetadataGateways, opts...)
	if err != nil {
		panic(err)
	}

	module := &Module{
		BaseModule: modules.New("contract_metadata_resolver"),
		pg:         pg,
		storage:    pg.Transactable,
		pool:       pool,
		cfg:        cfg,
		syncPeriod: time.Second * time.Duration(cfg.ContractMetadataResolver.SyncPeriod),
		retryDelay: time.Minute * time.Duration(cfg.ContractMetadataResolver.RetryDelay),
	}

	module.abiRegistry = &ABIRegistry{}
	if err = module.abiRegistry.LoadInterfaces(path.Join(cfg.Indexer.AssetsDir, "abi")); err != nil {
		panic(err)
	}

	return module
}

func (m *Module) Close() error {
	m.Log.Info().Msg("closing module...")
	m.G.Wait()

	return nil
}

func (m *Module) Start(ctx context.Context) {
	m.Log.Info().Msg("starting module...")
	m.G.GoCtx(ctx, m.receive)
}

func (m *Module) receive(ctx context.Context) {
	if err := m.sync(ctx); err != nil {
		m.Log.Err(err).Msg("sync")
	}

	ticker := time.NewTicker(m.syncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.sync(ctx); err != nil {
				m.Log.Err(err).Msg("sync")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Module) sync(ctx context.Context) error {
	cs, err := m.pg.Contracts.PendingMetadata(ctx, m.retryDelay, m.cfg.ContractMetadataResolver.RequestBulkSize)
	if err != nil {
		return errors.Wrap(err, "get contracts")
	}
	m.Log.Info().Int("count", len(cs)).Msg("new contracts received")

	contracts := make([]*storage.Contract, 0)
	sources := make([]*storage.Source, 0)
	for _, c := range cs {
		m.Log.Info().Str("contract", c.Address.String()).Str("metadata_link", c.MetadataLink).Msg("getting metadata...")
		metadata, err := m.pool.ContractMetadata(ctx, c.MetadataLink)
		if err != nil {
			m.failMetadata(c, err)
		} else {
			c.Language = metadata.Language
			c.OptimizerEnabled = metadata.Settings.Optimizer.Enabled
			c.ABI = metadata.Output.ABI

			if metadata.Output.ABI != nil {
				interfaces, err := m.abiRegistry.MatchABI(c.ABI)
				if err != nil {
					m.Log.Err(err).Str("contract", c.Address.String()).Any("ABI", c.ABI).Msg("failed ABI")
				}
				c.Tags = interfaces
			}

			c.Status = types.Success
			c.Error = ""

			contractSources, err := m.loadSources(ctx, c.Id, metadata.Sources)
			if err != nil {
				m.failMetadata(c, errors.Wrap(err, "get contract source metadata"))
			} else {
				sources = append(sources, contractSources...)
			}
		}

		contracts = append(contracts, c)
	}

	if len(contracts) == 0 {
		return nil
	}

	if err := m.save(ctx, contracts, sources); err != nil {
		return errors.Wrap(err, "saving contracts")
	}
	return nil
}

func (m *Module) loadSources(ctx context.Context, contractId uint64, metadataSources map[string]ipfs.Source) ([]*storage.Source, error) {
	if len(metadataSources) == 0 {
		return nil, nil
	}

	var (
		sources = make([]*storage.Source, 0, len(metadataSources))
		mu      sync.Mutex
	)

	g, groupCtx := errgroup.WithContext(ctx)
	for k, v := range metadataSources {
		source := &storage.Source{
			Name:       k,
			License:    v.License,
			Urls:       v.Urls,
			ContractId: contractId,
			Content:    v.Content,
		}
		if v.Content != "" {
			mu.Lock()
			sources = append(sources, source)
			mu.Unlock()
		} else {
			g.Go(func() error {
				content, err := m.pool.ContractText(groupCtx, v.Urls)
				if err != nil {
					return err
				}
				source.Content = content

				mu.Lock()
				sources = append(sources, source)
				mu.Unlock()

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return sources, nil
}

func (m *Module) save(ctx context.Context, contracts []*storage.Contract, sources []*storage.Source) error {
	tx, err := postgres.BeginTransaction(ctx, m.storage)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	_, err = tx.SaveContracts(ctx, contracts...)
	if err != nil {
		return errors.Wrap(err, "save contracts")
	}
	m.Log.Info().Int("count", len(contracts)).Msg("contract metadata updated")

	err = tx.SaveSources(ctx, sources...)
	if err != nil {
		return errors.Wrap(err, "save contract sources")
	}
	m.Log.Info().Int("count", len(sources)).Msg("contract sources saved")

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	return nil
}

func (m *Module) failMetadata(contract *storage.Contract, err error) {
	m.Log.Err(err).
		Str("contract", contract.Address.String()).
		Str("metadata_link", contract.MetadataLink).
		Msg(err.Error())

	contract.RetryCount += 1
	if contract.RetryCount >= m.cfg.ContractMetadataResolver.RetryCount {
		contract.Status = types.Failed
		m.Log.Err(err).Msg("retry limit exceeded. Status set to failed")
	}
	contract.Error = err.Error()
}

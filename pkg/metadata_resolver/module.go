package metadata_resolver

import (
	"context"
	"path"
	"time"

	"github.com/baking-bad/noble-indexer/internal/ipfs"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
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
	cfg         config.Config
	state       *storage.MetadataResolverState
	abiRegistry *ABIRegistry
}

func NewModule(pg postgres.Storage, cfg config.Config, opts ...ModuleOption) *Module {
	pool, err := ipfs.New(cfg.MetadataResolver.MetadataGateways)
	if err != nil {
		panic(err)
	}

	module := &Module{
		BaseModule: modules.New("metadata_scanner"),
		pg:         pg,
		storage:    pg.Transactable,
		pool:       pool,
		cfg:        cfg,
	}

	for i := range opts {
		opts[i](module)
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
	state, err := loadState(m.pg, m.storage, ctx, m.cfg.MetadataResolver.Name)
	if err != nil {
		m.Log.Err(err).Msg("while loading state")
		return
	}

	m.state = &state

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

func loadState(pg postgres.Storage, transactable sdk.Transactable, ctx context.Context, indexerName string) (storage.MetadataResolverState, error) {
	tx, err := postgres.BeginTransaction(ctx, transactable)
	if err != nil {
		return storage.MetadataResolverState{}, err
	}
	defer tx.Close(ctx)

	state, err := pg.MetadataResolverState.ByName(ctx, indexerName)
	if err != nil {
		if pg.State.IsNoRows(err) {
			state = storage.MetadataResolverState{
				Name:           indexerName,
				LastContractId: 0,
			}
			err = tx.Add(ctx, &state)
			if err != nil {
				return storage.MetadataResolverState{}, errors.Wrap(err, "save initial state")
			}

			if err = tx.Flush(ctx); err != nil {
				return storage.MetadataResolverState{}, tx.HandleError(ctx, err)
			}

		} else {
			return storage.MetadataResolverState{}, errors.Wrap(err, "get metadata resolver state")
		}
	}

	return state, nil
}

func (m *Module) sync(ctx context.Context) error {
	cs, err := m.pg.Contracts.ListWithMetadata(ctx, m.state.LastContractId)
	if err != nil {
		return errors.Wrap(err, "get contracts")
	}
	m.Log.Info().Int("count", len(cs)).Msg("new contracts received")

	contracts := make([]*storage.Contract, 0)
	sources := make([]*storage.Source, 0)
	for _, c := range cs {
		m.Log.Info().Str("contract", c.Address).Str("metadata_link", c.MetadataLink).Msg("getting metadata...")
		metadata, err := m.pool.ContractMetadata(ctx, c.MetadataLink)
		if err != nil {
			m.Log.Err(err).Str("contract", c.Address).Str("metadata_link", c.MetadataLink).Msg("failed to get metadata")
		}
		c.Language = metadata.Language
		c.OptimizerEnabled = metadata.Settings.Optimizer.Enabled
		c.ABI = metadata.Output.ABI

		if metadata.Output.ABI != nil {
			interfaces, err := m.abiRegistry.MatchABI(c.ABI)
			if err != nil {
				m.Log.Err(err).Str("contract", c.Address).Any("ABI", c.ABI).Msg("failed ABI")
			}
			c.Tags = interfaces
		}

		contracts = append(contracts, c)

		for k, v := range metadata.Sources {
			newSource := &storage.Source{
				Name:       k,
				License:    v.License,
				Urls:       v.Urls,
				ContractId: c.Id,
			}

			md, err := m.pool.ContractText(ctx, v.Urls)
			if err != nil {
				return errors.Wrap(err, "get contract source metadata")
			}
			newSource.Content = md
			sources = append(sources, newSource)
		}
	}

	if len(contracts) == 0 {
		return nil
	}

	tx, err := postgres.BeginTransaction(ctx, m.storage)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	err = tx.SaveContracts(ctx, contracts...)
	if err != nil {
		return errors.Wrap(err, "save contracts")
	}
	m.Log.Info().Int("count", len(contracts)).Msg("contract metadata updated")

	err = tx.SaveSources(ctx, sources...)
	if err != nil {
		return errors.Wrap(err, "save contract sources")
	}
	m.Log.Info().Int("count", len(sources)).Msg("contract sources saved")

	m.state.LastContractId = contracts[len(contracts)-1].Id
	err = tx.Update(ctx, m.state)
	if err != nil {
		return errors.Wrap(err, "update resolver state")
	}

	m.Log.Info().Uint64("last resolved contract id", contracts[len(contracts)-1].Id).Msg("resolver state updated")

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	return err
}

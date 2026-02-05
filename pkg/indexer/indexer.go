package indexer

import (
	"context"
	"sync"
	"time"

	internalStorage "github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/baking-bad/noble-indexer/pkg/indexer/genesis"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser"
	proxy "github.com/baking-bad/noble-indexer/pkg/indexer/proxy_contracts_resolver"
	"github.com/baking-bad/noble-indexer/pkg/indexer/receiver"
	"github.com/baking-bad/noble-indexer/pkg/indexer/rollback"
	"github.com/baking-bad/noble-indexer/pkg/indexer/storage"
	"github.com/baking-bad/noble-indexer/pkg/node"
	"github.com/baking-bad/noble-indexer/pkg/node/rpc"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	"github.com/dipdup-net/indexer-sdk/pkg/modules/stopper"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Indexer struct {
	cfg           config.Config
	api           node.Api
	receiver      *receiver.Module
	parser        *parser.Module
	proxyResolver *proxy.Module
	storage       *storage.Module
	genesis       *genesis.Module
	rollback      *rollback.Module
	stopper       modules.Module
	pg            postgres.Storage
	wg            *sync.WaitGroup
	log           zerolog.Logger
}

func New(ctx context.Context, cfg config.Config, stopperModule modules.Module) (Indexer, error) {
	pg, err := postgres.Create(ctx, cfg.Database, cfg.Indexer.ScriptsDir, true)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating pg context")
	}

	api, r, err := createReceiver(ctx, cfg, pg)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating receiver module")
	}

	networkConfig, err := cfg.Networks.Get(cfg.Network)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while getting network config")
	}

	p, err := createParser(cfg.Indexer, networkConfig, r)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating parser module")
	}

	proxyResolver, err := createProxyContractsResolver(cfg.Indexer, &api, &pg)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating parser module")
	}

	s, err := createStorage(pg, cfg, p, proxyResolver)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating storage module")
	}

	genesisModule, err := createGenesis(pg, cfg, r)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating genesis module")
	}

	rb, err := createRollback(r, pg, &api, cfg.Indexer)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating rollback module")
	}

	err = attachStopper(stopperModule, r, p, s, rb, genesisModule, proxyResolver)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating stopper module")
	}

	return Indexer{
		api:           &api,
		cfg:           cfg,
		receiver:      r,
		parser:        p,
		proxyResolver: proxyResolver,
		storage:       s,
		genesis:       genesisModule,
		rollback:      rb,
		stopper:       stopperModule,
		pg:            pg,
		wg:            new(sync.WaitGroup),
		log:           log.With().Str("module", "indexer").Logger(),
	}, nil
}

func (i *Indexer) Start(ctx context.Context) {
	i.log.Info().Msg("starting...")

	i.genesis.Start(ctx)
	i.storage.Start(ctx)
	i.proxyResolver.Start(ctx)
	i.parser.Start(ctx)
	i.rollback.Start(ctx)
	i.receiver.Start(ctx)
}

func (i *Indexer) Close() error {
	i.log.Info().Msg("closing...")
	i.wg.Wait()

	if err := i.receiver.Close(); err != nil {
		log.Err(err).Msg("closing receiver")
	}
	if err := i.genesis.Close(); err != nil {
		log.Err(err).Msg("closing genesis")
	}
	if err := i.parser.Close(); err != nil {
		log.Err(err).Msg("closing parser")
	}
	if err := i.proxyResolver.Close(); err != nil {
		log.Err(err).Msg("closing proxy contracts resolver")
	}
	if err := i.storage.Close(); err != nil {
		log.Err(err).Msg("closing storage")
	}
	if err := i.rollback.Close(); err != nil {
		log.Err(err).Msg("closing rollback")
	}
	if err := i.pg.Close(); err != nil {
		log.Err(err).Msg("closing postgres connection")
	}

	return nil
}

func createReceiver(ctx context.Context, cfg config.Config, pg postgres.Storage) (rpc.API, *receiver.Module, error) {
	state, err := loadState(pg, ctx, cfg.Indexer.Name)
	if err != nil {
		return rpc.API{}, nil, errors.Wrap(err, "while loading state")
	}

	var (
		ws      *websocket.Conn
		nodeRpc rpc.API
	)

	if ds, ok := cfg.DataSources["node_rpc"]; ok && ds.URL != "" {
		nodeRpc = rpc.NewApi(ds, rpc.WithTimeout(time.Second*time.Duration(ds.Timeout)), rpc.WithRateLimit(ds.RequestsPerSecond))
	}
	if ds, ok := cfg.DataSources["node_ws"]; ok && ds.URL != "" && ds.Credentials.ApiKey != nil {
		ws, _, err = websocket.DefaultDialer.Dial(ds.URL, nil)
		if err != nil {
			return nodeRpc, nil, errors.Wrap(err, "create websocket")
		}
	}

	receiverModule := receiver.NewModule(cfg.Indexer, &nodeRpc, ws, state)
	return nodeRpc, &receiverModule, nil
}

func createParser(cfg config.Indexer, networkConfig config.Network, receiverModule modules.Module) (*parser.Module, error) {
	parserModule := parser.NewModule(cfg, networkConfig)

	if err := parserModule.AttachTo(receiverModule, receiver.BlocksOutput, parser.InputName); err != nil {
		return nil, errors.Wrap(err, "while attaching parser to receiver")
	}

	return &parserModule, nil
}

func createProxyContractsResolver(cfg config.Indexer, api *rpc.API, pg *postgres.Storage) (*proxy.Module, error) {
	proxyContractsResolver := proxy.NewModule(cfg, api, pg)
	return &proxyContractsResolver, nil
}

func createStorage(
	pg postgres.Storage,
	cfg config.Config,
	parserModule modules.Module,
	proxyResolverModule modules.Module,
) (*storage.Module, error) {
	storageModule := storage.NewModule(pg, pg.Notificator, cfg.Indexer)

	if err := storageModule.AttachTo(parserModule, parser.OutputName, storage.InputName); err != nil {
		return nil, errors.Wrap(err, "while attaching storage to parser")
	}

	if err := storageModule.AttachTo(
		proxyResolverModule,
		proxy.OutputName,
		storage.ProxyContractsInput,
	); err != nil {
		return nil, errors.Wrap(err, "while attaching storage to proxy contracts resolver")
	}

	return &storageModule, nil
}

func createGenesis(pg postgres.Storage, cfg config.Config, receiverModule modules.Module) (*genesis.Module, error) {
	genesisModule := genesis.NewModule(pg, cfg.Indexer)

	if err := genesisModule.AttachTo(receiverModule, receiver.GenesisOutput, genesis.InputName); err != nil {
		return nil, errors.Wrap(err, "while attaching genesis to receiver")
	}

	genesisModulePtr := &genesisModule
	if err := receiverModule.AttachTo(genesisModulePtr, genesis.OutputName, receiver.GenesisDoneInput); err != nil {
		return nil, errors.Wrap(err, "while attaching receiver to genesis")
	}

	return genesisModulePtr, nil
}

func createRollback(receiverModule modules.Module, pg postgres.Storage, api node.Api, cfg config.Indexer) (*rollback.Module, error) {
	rollbackModule := rollback.NewModule(pg.Transactable, pg.State, pg.Blocks, api, cfg)

	// rollback <- listen signal -- receiver
	if err := rollbackModule.AttachTo(receiverModule, receiver.RollbackOutput, rollback.InputName); err != nil {
		return nil, errors.Wrap(err, "while attaching rollback to receiver")
	}

	// receiver <- listen state -- rollback
	if err := receiverModule.AttachTo(&rollbackModule, rollback.OutputName, receiver.RollbackInput); err != nil {
		return nil, errors.Wrap(err, "while attaching receiver to rollback")
	}

	return &rollbackModule, nil
}

func attachStopper(
	stopperModule modules.Module,
	receiverModule modules.Module,
	parserModule modules.Module,
	storageModule modules.Module,
	rollbackModule modules.Module,
	genesisModule modules.Module,
	proxyResolverModule modules.Module,
) error {
	if err := stopperModule.AttachTo(receiverModule, receiver.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to receiver")
	}

	if err := stopperModule.AttachTo(parserModule, parser.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to parser")
	}

	if err := stopperModule.AttachTo(storageModule, storage.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to storage")
	}

	if err := stopperModule.AttachTo(rollbackModule, rollback.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to rollback")
	}

	if err := stopperModule.AttachTo(genesisModule, genesis.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to genesis")
	}

	if err := stopperModule.AttachTo(proxyResolverModule, proxy.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to proxy contracts resolver")
	}

	return nil
}

func loadState(pg postgres.Storage, ctx context.Context, indexerName string) (*internalStorage.State, error) {
	state, err := pg.State.ByName(ctx, indexerName)
	if err != nil {
		if pg.State.IsNoRows(err) {
			return nil, nil
		}

		return nil, err
	}

	return &state, nil
}

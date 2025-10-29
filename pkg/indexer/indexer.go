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
	"github.com/baking-bad/noble-indexer/pkg/indexer/receiver"
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
	cfg      config.Config
	api      node.Api
	receiver *receiver.Module
	parser   *parser.Module
	storage  *storage.Module
	genesis  *genesis.Module
	stopper  modules.Module
	wg       *sync.WaitGroup
	log      zerolog.Logger
}

func New(ctx context.Context, cfg config.Config, stopperModule modules.Module) (Indexer, error) {
	pg, err := postgres.Create(ctx, cfg.Database, cfg.Indexer.ScriptsDir)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating pg context")
	}

	api, r, err := createReceiver(ctx, cfg, pg)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating receiver module")
	}

	p, err := createParser(cfg.Indexer, r)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating parser module")
	}

	s, err := createStorage(pg, cfg, p)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating storage module")
	}

	genesisModule, err := createGenesis(pg, cfg, r)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating genesis module")
	}

	err = attachStopper(stopperModule, r, p, s, genesisModule)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating stopper module")
	}

	return Indexer{
		api:      &api,
		cfg:      cfg,
		receiver: r,
		parser:   p,
		storage:  s,
		genesis:  genesisModule,
		stopper:  stopperModule,
		wg:       new(sync.WaitGroup),
		log:      log.With().Str("module", "indexer").Logger(),
	}, nil
}

func (i *Indexer) Start(ctx context.Context) {
	i.log.Info().Msg("starting...")

	i.genesis.Start(ctx)
	i.storage.Start(ctx)
	i.parser.Start(ctx)
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
	if err := i.storage.Close(); err != nil {
		log.Err(err).Msg("closing storage")
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

func createParser(cfg config.Indexer, receiverModule modules.Module) (*parser.Module, error) {
	parserModule := parser.NewModule(cfg)

	if err := parserModule.AttachTo(receiverModule, receiver.BlocksOutput, parser.InputName); err != nil {
		return nil, errors.Wrap(err, "while attaching parser to receiver")
	}

	return &parserModule, nil
}

func createStorage(pg postgres.Storage, cfg config.Config, parserModule modules.Module) (*storage.Module, error) {
	storageModule := storage.NewModule(pg, cfg.Indexer)

	if err := storageModule.AttachTo(parserModule, parser.OutputName, storage.InputName); err != nil {
		return nil, errors.Wrap(err, "while attaching storage to parser")
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

func attachStopper(
	stopperModule modules.Module,
	receiverModule modules.Module,
	parserModule modules.Module,
	storageModule modules.Module,
	genesisModule modules.Module,
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

	if err := stopperModule.AttachTo(genesisModule, genesis.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to genesis")
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

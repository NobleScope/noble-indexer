package indexer

import (
	"context"
	"sync"
	"time"

	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/baking-bad/noble-indexer/pkg/indexer/receiver"
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

	stopper modules.Module
	wg      *sync.WaitGroup
	log     zerolog.Logger
}

func New(ctx context.Context, cfg config.Config, stopperModule modules.Module) (Indexer, error) {
	api, r, err := createReceiver(ctx, cfg)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating receiver module")
	}

	err = attachStopper(stopperModule, r)
	if err != nil {
		return Indexer{}, errors.Wrap(err, "while creating stopper module")
	}

	return Indexer{
		api:      &api,
		cfg:      cfg,
		receiver: r,
		stopper:  stopperModule,
		wg:       new(sync.WaitGroup),
		log:      log.With().Str("module", "indexer").Logger(),
	}, nil
}

func (i *Indexer) Start(ctx context.Context) {
	i.log.Info().Msg("starting...")

	i.receiver.Start(ctx)
}

func (i *Indexer) Close() error {
	i.log.Info().Msg("closing...")
	i.wg.Wait()

	if err := i.receiver.Close(); err != nil {
		log.Err(err).Msg("closing receiver")
	}

	return nil
}

func createReceiver(ctx context.Context, cfg config.Config) (rpc.API, *receiver.Module, error) {
	var (
		err     error
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

	receiverModule := receiver.NewModule(cfg.Indexer, &nodeRpc, ws)
	return nodeRpc, &receiverModule, nil
}

func attachStopper(
	stopperModule modules.Module,
	receiverModule modules.Module,
) error {
	if err := stopperModule.AttachTo(receiverModule, receiver.StopOutput, stopper.InputName); err != nil {
		return errors.Wrap(err, "while attaching stopper to receiver")
	}

	return nil
}

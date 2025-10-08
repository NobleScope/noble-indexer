package parser

import (
	"context"

	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
)

type Module struct {
	modules.BaseModule

	cfg config.Indexer
}

var _ modules.Module = (*Module)(nil)

const (
	InputName  = "blocks"
	OutputName = "data"
	StopOutput = "stop"
)

func NewModule(cfg config.Indexer) Module {
	m := Module{
		BaseModule: modules.New("parser"),
		cfg:        cfg,
	}
	m.CreateInputWithCapacity(InputName, 128)
	m.CreateOutput(OutputName)
	m.CreateOutput(StopOutput)

	return m
}

func (p *Module) Start(ctx context.Context) {
	p.Log.Info().Msg("starting parser module...")
	p.G.GoCtx(ctx, p.listen)
}

func (p *Module) Close() error {
	p.Log.Info().Msg("closing...")
	p.G.Wait()
	return nil
}

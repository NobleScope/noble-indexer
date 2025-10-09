package parser

import (
	"time"

	"github.com/baking-bad/noble-indexer/pkg/types"
)

func (p *Module) parse(b types.BlockData) error {
	start := time.Now()
	height, err := b.Number.Uint64()
	if err != nil {
		return err
	}
	p.Log.Info().
		Uint64("height", height).
		Msg("parsing block...")

	p.Log.Info().
		Uint64("height", height).
		Int64("ms", time.Since(start).Milliseconds()).
		Msg("block parsed")

	output := p.MustOutput(OutputName)
	output.Push(b)

	return nil
}

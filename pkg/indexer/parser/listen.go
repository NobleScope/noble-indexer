package parser

import (
	"context"

	"github.com/NobleScope/noble-indexer/pkg/types"
)

func (p *Module) listen(ctx context.Context) {
	p.Log.Info().Msg("module started")

	input := p.MustInput(InputName)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-input.Listen():
			if !ok {
				p.Log.Warn().Msg("can't read message from input, it was drained and closed")
				p.MustOutput(StopOutput).Push(struct{}{})
				return
			}

			block, ok := msg.(types.BlockData)
			if !ok {
				p.Log.Warn().Msgf("invalid message type: %T", msg)
				continue
			}

			if parseErr := p.parse(block); parseErr != nil {
				height, err := block.Number.Uint64()
				if err != nil {
					p.Log.Warn().Err(err).Str("num", block.Number.String()).Msg("can't parse block number")
				}
				p.Log.Err(parseErr).
					Uint64("height", height).
					Msg("block parsing error")
				p.MustOutput(StopOutput).Push(struct{}{})
				continue
			}
		}
	}
}

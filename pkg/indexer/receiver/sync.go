package receiver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/baking-bad/noble-indexer/pkg/node/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/pkg/errors"
)

func (r *Module) sync(ctx context.Context) {
	var blocksCtx context.Context
	blocksCtx, r.cancelReadBlocks = context.WithCancel(ctx)
	if err := r.readBlocks(blocksCtx); err != nil {
		r.Log.Err(err).Msg("while reading blocks")
		r.stopAll()
		return
	}

	if ctx.Err() != nil {
		return
	}

	if r.ws != nil {
		if err := r.live(ctx); err != nil {
			r.Log.Err(err).Msg("while reading blocks")
			r.stopAll()
			return
		}
	} else {
		ticker := time.NewTicker(time.Second * time.Duration(r.cfg.BlockPeriod))
		defer ticker.Stop()

		for {
			r.rollbackSync.Wait()

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				blocksCtx, r.cancelReadBlocks = context.WithCancel(ctx)
				if err := r.readBlocks(blocksCtx); err != nil && !errors.Is(err, context.Canceled) {
					r.Log.Err(err).Msg("while reading blocks by timer")
					r.stopAll()
					return
				}
			}
		}
	}
}

func (r *Module) live(ctx context.Context) error {
	subscibeReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "eth_subscribe",
		"params":  []interface{}{"newHeads"},
	}

	if err := r.ws.WriteJSON(subscibeReq); err != nil {
		return err
	}

	r.Log.Info().Msg("websocket was subscribed on block header events")

	for {
		r.rollbackSync.Wait()

		select {
		case <-ctx.Done():
			return nil
		default:
			_, msg, err := r.ws.ReadMessage()
			if err != nil {
				return err
			}

			var wsMsg types.Response[pkgTypes.Block]
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				r.Log.Err(err).Any("Raw message:", string(msg)).Msg("failed to unmarshal ws message")
				continue
			}

			if wsMsg.Method != "eth_subscription" {
				continue
			}

			var payload pkgTypes.Payload
			if err := json.Unmarshal(wsMsg.Params, &payload); err != nil {
				r.Log.Err(err).Any("Unmarshal params:", wsMsg.Params).Msg("failed to unmarshal subscription params")
				continue
			}

			if payload.Result == nil {
				continue
			}

			height, err := payload.Result.Number.Uint64()
			if err != nil {
				r.Log.Err(err).Msg("failed to parse block number")
				continue
			}

			r.Log.Info().Uint64("height", height).Msg("ws subscription received")
			r.passBlocks(ctx, pkgTypes.Level(height))
		}
	}
}

func (r *Module) readBlocks(ctx context.Context) error {
	for {
		headLevel, err := r.headLevel(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		isLiveMode := headLevel-r.level < pkgTypes.Level(r.w.capacity)
		r.w.SetLiveMode(isLiveMode)

		if level, _ := r.Level(); level == headLevel {
			time.Sleep(time.Millisecond * 300)
			continue
		}

		r.passBlocks(ctx, headLevel)
		return nil
	}
}

func (r *Module) passBlocks(ctx context.Context, head pkgTypes.Level) {
	level, _ := r.Level()
	level += 1

	for ; level <= head; level++ {
		select {
		case <-ctx.Done():
			return
		default:
			r.w.Do(ctx, level)
		}
	}
}

func (r *Module) headLevel(ctx context.Context) (pkgTypes.Level, error) {
	head, err := r.api.Head(ctx)
	if err != nil {
		return 0, err
	}
	return head, nil
}

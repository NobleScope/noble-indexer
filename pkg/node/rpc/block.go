package rpc

import (
	"context"
	stdjson "encoding/json"
	"net/url"

	"github.com/NobleScope/noble-indexer/pkg/node/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/opus-domini/fast-shot/constant/header"
	"github.com/pkg/errors"
)

const (
	pathBlock    = "eth_getBlockByNumber"
	pathReceipts = "eth_getBlockReceipts"
)

func (api *API) Block(ctx context.Context, level pkgTypes.Level) (pkgTypes.Block, error) {
	if api.rateLimit != nil {
		if err := api.rateLimit.Wait(ctx); err != nil {
			return pkgTypes.Block{}, err
		}
	}

	u, err := url.Parse(api.cfg.URL)
	if err != nil {
		return pkgTypes.Block{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, api.timeout)
	defer cancel()

	hexLevel := level.Hex()

	request := types.Request{
		Method:  pathBlock,
		JsonRpc: "2.0",
		Id:      1,
		Params: []any{
			hexLevel,
			true,
		},
	}

	resp, err := api.client.POST(u.Path).
		Context().
		Set(requestCtx).
		Header().
		AddAll(map[header.Type]string{
			header.ContentType: "application/json",
			header.UserAgent:   userAgent}).
		Body().AsJSON(&request).Send()
	if err != nil {
		return pkgTypes.Block{}, err
	}

	if resp.Status().IsError() {
		return pkgTypes.Block{}, errors.Errorf("invalid status: %d", resp.Status().Code())
	}

	var block types.Response[pkgTypes.Block]
	err = json.NewDecoder(resp.Raw().Body).Decode(&block)

	return block.Result, err
}

func (api *API) BlockBulk(ctx context.Context, levels ...pkgTypes.Level) ([]pkgTypes.BlockData, error) {
	if len(levels) == 0 {
		return nil, nil
	}
	if api.rateLimit != nil {
		if err := api.rateLimit.Wait(ctx); err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(api.cfg.URL)
	if err != nil {
		return []pkgTypes.BlockData{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, api.timeout)
	defer cancel()

	requests := make([]types.Request, len(levels)*3)

	for i := range levels {
		hexLevel := levels[i].Hex()
		requests[i*3] = types.Request{
			Method:  pathBlock,
			JsonRpc: "2.0",
			Id:      int64(i * 3),
			Params: []any{
				hexLevel,
				true,
			},
		}
		requests[i*3+1] = types.Request{
			Method:  pathReceipts,
			JsonRpc: "2.0",
			Id:      int64(i*3 + 1),
			Params: []any{
				hexLevel,
			},
		}

		requests[i*3+2] = types.Request{
			Method:  api.traceProvider.Method(),
			JsonRpc: "2.0",
			Id:      int64(i*3 + 2),
			Params:  api.traceProvider.Params(hexLevel),
		}
	}

	resp, err := api.client.POST(u.Path).
		Context().
		Set(requestCtx).
		Header().
		AddAll(map[header.Type]string{
			header.ContentType: "application/json",
			header.UserAgent:   userAgent}).
		Body().AsJSON(&requests).Send()

	if err != nil {
		return []pkgTypes.BlockData{}, err
	}
	if resp.Status().IsError() {
		return []pkgTypes.BlockData{}, errors.Errorf("invalid status: %d", resp.Status().Code())
	}

	var rawResponses []types.Response[stdjson.RawMessage]
	err = json.NewDecoder(resp.Raw().Body).Decode(&rawResponses)
	if err != nil {
		return []pkgTypes.BlockData{}, err
	}

	var blockData = make([]pkgTypes.BlockData, len(levels))
	for i := range rawResponses {
		if rawResponses[i].Error != nil {
			return nil, errors.Wrapf(types.ErrRequest, "request error: %s", rawResponses[i].Error.Error())
		}

		blockIdx := rawResponses[i].Id / 3
		switch rawResponses[i].Id % 3 {
		case 0:
			var block pkgTypes.Block
			if err := json.Unmarshal(rawResponses[i].Result, &block); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal block")
			}
			blockData[blockIdx].Block = block
		case 1:
			if len(rawResponses[i].Result) == 0 {
				blockData[blockIdx].Receipts = []pkgTypes.Receipt{}
				continue
			}
			var receipts []pkgTypes.Receipt
			if err := json.Unmarshal(rawResponses[i].Result, &receipts); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal receipts")
			}
			blockData[blockIdx].Receipts = receipts
		case 2:
			if len(rawResponses[i].Result) == 0 {
				blockData[blockIdx].Traces = []pkgTypes.Trace{}
				continue
			}
			traces, err := api.traceProvider.ParseTraces(rawResponses[i].Result)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse traces")
			}
			blockData[blockIdx].Traces = traces
		}
	}

	return blockData, nil
}

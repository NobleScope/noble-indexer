package rpc

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/baking-bad/noble-indexer/pkg/node/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/pkg/errors"
)

const (
	pathBlock    = "eth_getBlockByNumber"
	pathReceipts = "eth_getBlockReceipts"
)

func (api *API) Block(ctx context.Context, level pkgTypes.Level) (pkgTypes.Block, error) {
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
		Context().Set(requestCtx).
		Header().Add("User-Agent", userAgent).
		Body().AsJSON(&request).
		Send()
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

	u, err := url.Parse(api.cfg.URL)
	if err != nil {
		return []pkgTypes.BlockData{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, api.timeout)
	defer cancel()

	responses := make([]any, len(levels)*2)
	requests := make([]types.Request, len(levels)*2)

	for i := range levels {
		responses[i*2] = &types.Response[pkgTypes.Block]{}
		responses[i*2+1] = &types.Response[[]pkgTypes.Receipt]{}

		hexLevel := levels[i].Hex()
		requests[i*2] = types.Request{
			Method:  pathBlock,
			JsonRpc: "2.0",
			Id:      1,
			Params: []any{
				hexLevel,
				true,
			},
		}
		requests[i*2+1] = types.Request{
			Method:  pathReceipts,
			JsonRpc: "2.0",
			Id:      1,
			Params: []any{
				hexLevel,
			},
		}
	}

	resp, err := api.client.POST(u.Path).
		Context().Set(requestCtx).
		Header().Add("User-Agent", userAgent).
		Body().AsJSON(&requests).
		Send()
	if err != nil {
		return []pkgTypes.BlockData{}, err
	}
	if resp.Status().IsError() {
		return []pkgTypes.BlockData{}, errors.Errorf("invalid status: %d", resp.Status().Code())
	}

	err = json.NewDecoder(resp.Raw().Body).Decode(&responses)
	if err != nil {
		return []pkgTypes.BlockData{}, err
	}

	var blockData = make([]pkgTypes.BlockData, len(levels))
	for i := range responses {
		switch typ := responses[i].(type) {
		case *types.Response[pkgTypes.Block]:
			if typ.Error != nil {
				return nil, errors.Wrapf(types.ErrRequest, "request error: %s", typ.Error.Error())
			}
			blockData[i/2].Block = typ.Result
		case *types.Response[[]pkgTypes.Receipt]:
			if typ.Error != nil {
				return nil, errors.Wrapf(types.ErrRequest, "request error: %s", typ.Error.Error())
			}
			blockData[i/2].Receipts = typ.Result
		}
	}

	return blockData, nil
}

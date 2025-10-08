package rpc

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/baking-bad/noble-indexer/pkg/node/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/pkg/errors"
)

const pathHead = "eth_blockNumber"

func (api *API) Head(ctx context.Context) (pkgTypes.Level, error) {
	u, err := url.Parse(api.cfg.URL)
	if err != nil {
		return 0, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, api.timeout)
	defer cancel()

	request := types.Request{
		Method:  pathHead,
		JsonRpc: "2.0",
		Id:      1,
		Params:  []any{},
	}

	resp, err := api.client.POST(u.Path).
		Context().Set(requestCtx).
		Header().Add("User-Agent", userAgent).
		Body().AsJSON(&request).
		Send()
	if err != nil {
		return 0, err
	}

	if resp.Status().IsError() {
		return 0, errors.Errorf("invalid status: %d", resp.Status().Code())
	}

	results := types.Response[pkgTypes.Hex]{}
	err = json.NewDecoder(resp.Raw().Body).Decode(&results)

	val, err := results.Result.Uint64()
	if err != nil {
		panic(err)
	}

	api.log.Debug().Uint64("head", val)

	return pkgTypes.Level(val), err
}

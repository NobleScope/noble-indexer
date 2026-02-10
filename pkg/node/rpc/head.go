package rpc

import (
	"context"
	"net/url"

	"github.com/NobleScope/noble-indexer/pkg/node/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/opus-domini/fast-shot/constant/header"
	"github.com/pkg/errors"
)

const pathHead = "eth_blockNumber"

func (api *API) Head(ctx context.Context) (pkgTypes.Level, error) {
	if api.rateLimit != nil {
		if err := api.rateLimit.Wait(ctx); err != nil {
			return 0, err
		}
	}

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
		Context().
		Set(requestCtx).
		Header().
		AddAll(map[header.Type]string{
			header.ContentType: "application/json",
			header.UserAgent:   userAgent}).
		Body().AsJSON(&request).Send()

	if err != nil {
		return 0, err
	}

	if resp.Status().IsError() {
		return 0, errors.Errorf("invalid status: %d", resp.Status().Code())
	}

	results := types.Response[pkgTypes.Hex]{}
	err = json.NewDecoder(resp.Raw().Body).Decode(&results)
	if err != nil {
		return 0, errors.Wrapf(err, "decoding response body")
	}

	val, err := results.Result.Uint64()
	if err != nil {
		api.log.Err(err).Msg("converting level")
		panic(err)
	}

	api.log.Debug().Uint64("head", val).Send()

	return pkgTypes.Level(val), err
}

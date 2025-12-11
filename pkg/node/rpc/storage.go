package rpc

import (
	"context"
	"net/url"

	"github.com/opus-domini/fast-shot/constant/header"

	"github.com/baking-bad/noble-indexer/pkg/node/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/pkg/errors"
)

const pathStorage = "eth_getStorageAt"

func (api *API) Storage(ctx context.Context, requestData []pkgTypes.StorageRequest) ([]pkgTypes.Hex, error) {
	u, err := url.Parse(api.cfg.URL)
	if err != nil {
		return nil, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, api.timeout)
	defer cancel()

	params := make([]types.Request, len(requestData))
	for i := range requestData {
		params[i] = types.Request{
			Method:  pathStorage,
			JsonRpc: "2.0",
			Id:      int64(i + 1),
			Params: []any{
				requestData[i].ContractAddress,
				requestData[i].StorageSlot,
				requestData[i].BlockNumber,
			},
		}
	}

	resp, err := api.client.POST(u.Path).
		Context().
		Set(requestCtx).
		Header().
		AddAll(map[header.Type]string{
			header.ContentType: "application/json",
			header.UserAgent:   userAgent}).
		Body().AsJSON(&params).Send()
	if err != nil {
		return nil, err
	}

	if resp.Status().IsError() {
		return nil, errors.Errorf("invalid status: %d", resp.Status().Code())
	}

	var response []types.Response[pkgTypes.Hex]
	if err = json.NewDecoder(resp.Raw().Body).Decode(&response); err != nil {
		return nil, err
	}

	result := make([]pkgTypes.Hex, len(response))
	for i := range response {
		result[i] = response[i].Result
	}

	return result, err
}

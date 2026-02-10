package rpc

import (
	"context"
	"fmt"
	"net/url"

	internalTypes "github.com/NobleScope/noble-indexer/internal/storage/types"
	"github.com/NobleScope/noble-indexer/pkg/node/types"
	tmTypes "github.com/NobleScope/noble-indexer/pkg/token_metadata/types"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opus-domini/fast-shot/constant/header"
)

const pathEthCall = "eth_call"

type TokenCall struct {
	Query  pkgTypes.TokenMetadataRequest
	Method tmTypes.TokenEndpoint
}

func (api *API) TokenMetadataBulk(
	ctx context.Context,
	queries []pkgTypes.TokenMetadataRequest,
) (map[uint64]pkgTypes.TokenMetadata, error) {
	if api.rateLimit != nil {
		if err := api.rateLimit.Wait(ctx); err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(api.cfg.URL)
	if err != nil {
		return nil, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, api.timeout)
	defer cancel()

	var calls []TokenCall
	for _, q := range queries {
		switch q.Interface {
		case internalTypes.ERC20:
			calls = append(calls,
				TokenCall{q, tmTypes.Name},
				TokenCall{q, tmTypes.Symbol},
				TokenCall{q, tmTypes.Decimals},
			)

		case internalTypes.ERC721:
			calls = append(calls,
				TokenCall{q, tmTypes.Name},
				TokenCall{q, tmTypes.Symbol},
				TokenCall{q, tmTypes.TokenUri},
			)

		case internalTypes.ERC1155:
			if q.TokenID == nil {
				return nil, fmt.Errorf("ERC1155 query missing TokenID for contract %s", q.Address)
			}
			calls = append(calls,
				TokenCall{q, tmTypes.Uri},
			)

		default:
			return nil, fmt.Errorf("unknown interface: %s", q.Interface)
		}
	}

	requests := make([]types.Request, 0, len(calls))
	idToCall := make(map[int64]TokenCall, len(calls))
	for i, call := range calls {
		packed, err := packCall(call)
		if err != nil {
			return nil, err
		}

		reqID := int64(i) + 1
		req := types.Request{
			JsonRpc: "2.0",
			Id:      reqID,
			Method:  pathEthCall,
			Params: []interface{}{
				map[string]interface{}{
					"to":   common.HexToAddress(call.Query.Address).Hex(),
					"data": "0x" + common.Bytes2Hex(packed),
				},
				"latest",
			},
		}

		requests = append(requests, req)
		idToCall[reqID] = call
	}

	var responses []types.Response[pkgTypes.Hex]
	resp, err := api.client.POST(u.Path).
		Context().
		Set(requestCtx).
		Header().
		AddAll(map[header.Type]string{
			header.ContentType: "application/json",
			header.UserAgent:   userAgent,
		}).
		Body().AsJSON(&requests).Send()

	if err != nil {
		return nil, err
	}
	if resp.Status().IsError() {
		return nil, fmt.Errorf("invalid status: %d", resp.Status().Code())
	}

	err = json.NewDecoder(resp.Raw().Body).Decode(&responses)
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]pkgTypes.TokenMetadata)
	for _, r := range responses {
		call, ok := idToCall[r.Id]
		if !ok {
			return nil, fmt.Errorf("response ID %d not found in token call map", r.Id)
		}
		raw := r.Result.Bytes()

		md := result[call.Query.Id]
		switch call.Query.Interface {
		case internalTypes.ERC20:
			switch call.Method {
			case tmTypes.Name:
				md.Name = raw
			case tmTypes.Symbol:
				md.Symbol = raw
			case tmTypes.Decimals:
				md.Decimals = raw
			default:
				continue
			}

		case internalTypes.ERC721:
			switch call.Method {
			case tmTypes.Name:
				md.Name = raw
			case tmTypes.Symbol:
				md.Symbol = raw
			case tmTypes.TokenUri:
				md.URI = raw
			default:
				continue
			}

		case internalTypes.ERC1155:
			if call.Method == tmTypes.Uri {
				md.URI = raw
			}
		}

		result[call.Query.Id] = md
	}

	return result, nil
}

func packCall(c TokenCall) ([]byte, error) {
	switch c.Query.Interface {
	case internalTypes.ERC20:
		return c.Query.ABI.Pack(c.Method.String())

	case internalTypes.ERC1155:
		return c.Query.ABI.Pack(c.Method.String(), c.Query.TokenID)

	case internalTypes.ERC721:
		if c.Method == tmTypes.TokenUri {
			return c.Query.ABI.Pack(c.Method.String(), c.Query.TokenID)
		}
		return c.Query.ABI.Pack(c.Method.String())

	default:
		return nil, fmt.Errorf("unknown interface: %s", c.Query.Interface)
	}
}

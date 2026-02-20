package rpc

import (
	"time"

	"github.com/NobleScope/noble-indexer/pkg/node/rpc/trace_provider"
	"golang.org/x/time/rate"
)

type APIOption func(api *API)

func WithRateLimit(rps int) APIOption {
	return func(api *API) {
		api.rateLimit = rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), rps)
	}
}

func WithTimeout(timeout time.Duration) APIOption {
	return func(api *API) {
		api.timeout = timeout
	}
}

func WithTraceMethod(method string) APIOption {
	return func(api *API) {
		switch method {
		case "debug_traceBlockByNumber":
			api.traceProvider = &trace_provider.GethDebugTraceProvider{}
		default:
			api.traceProvider = &trace_provider.ParityTraceProvider{}
		}
	}
}

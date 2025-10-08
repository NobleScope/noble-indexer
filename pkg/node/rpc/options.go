package rpc

import (
	"time"

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

package cache

import (
	"context"
	"io"
	"time"
)

// ICache is an interface for caching data with expiration.
// It provides methods to get and set data, as well as to close the cache when it's no longer needed.
//
//go:generate mockgen -source=$GOFILE -destination=mock.go -package=cache -typed
type ICache interface {
	io.Closer

	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key string, data string, f ExpirationFunc) error
}

type ExpirationFunc func() time.Duration

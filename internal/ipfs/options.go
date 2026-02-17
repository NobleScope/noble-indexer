package ipfs

import "github.com/NobleScope/noble-indexer/internal/cache"

type Option func(*Pool)

func WithCache(cache cache.ICache) Option {
	return func(i *Pool) {
		i.cache = cache
	}
}

package ipfs

import (
	"context"
	"testing"
	"time"

	"github.com/baking-bad/noble-indexer/internal/cache"
	"github.com/dipdup-io/ipfs-tools"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Pool_LoadMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	ipfsMock := ipfs.NewMockIPool(ctrl)

	mockCache := cache.NewMockICache(ctrl)

	pool := &Pool{
		ipfs:  ipfsMock,
		cache: mockCache,
	}
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	data := []byte(`{"name":"ipfs","description":"IPFS is a peer-to-peer hypermedia protocol designed to make the web faster, safer, and more open.","image":"ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"}`)

	ipfsMock.EXPECT().
		Get(gomock.Any(), ipfs.Path("ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG")).
		Return(ipfs.Data{
			Raw:          data,
			ResponseTime: 10,
			Node:         "node-1",
		}, nil).
		Times(1)

	mockCache.EXPECT().
		Get(gomock.Any(), "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG").
		Return("", false).
		Times(1)

	mockCache.EXPECT().
		Set(gomock.Any(), "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", string(data), nil).
		Return(nil).
		Times(1)

	response, err := pool.LoadMetadata(ctx, "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG")
	require.NoError(t, err)
	require.Equal(t, data, response)

	mockCache.EXPECT().
		Get(gomock.Any(), "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG").
		Return(string(data), true).
		Times(1)

	// from cache
	response, err = pool.LoadMetadata(ctx, "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG")
	require.NoError(t, err)
	require.Equal(t, data, response)
}

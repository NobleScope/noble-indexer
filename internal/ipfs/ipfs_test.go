package ipfs

import (
	"context"
	"testing"
	"time"

	"github.com/dipdup-io/ipfs-tools"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Pool_LoadMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	ipfsMock := ipfs.NewMockIPool(ctrl)

	cache, err := lru.New[string, []byte](2)
	require.NoError(t, err)

	pool := &Pool{
		ipfs:  ipfsMock,
		cache: cache,
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

	response, err := pool.LoadMetadata(ctx, "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG")
	require.NoError(t, err)
	require.Equal(t, data, response)

	// from cache
	response, err = pool.LoadMetadata(ctx, "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG")
	require.NoError(t, err)
	require.Equal(t, data, response)
}

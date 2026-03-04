package rpc

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	indexerConfig "github.com/NobleScope/noble-indexer/pkg/indexer/config"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/dipdup-net/go-lib/config"
	"gopkg.in/yaml.v3"
)

const requestBulkSize = 50

var partialCfg struct {
	Networks indexerConfig.NetworksConfig `yaml:"networks"`
}

func newBenchAPI(tb testing.TB) API {
	tb.Helper()

	nodeURL := os.Getenv("EVM_NODE_URL")
	if nodeURL == "" {
		tb.Skip("EVM_NODE_URL is not set")
	}

	network := os.Getenv("NETWORK")
	if network == "" {
		tb.Skip("NETWORK is not set")
	}

	raw, err := os.ReadFile("../../../configs/dipdup.yml")
	if err != nil {
		tb.Fatalf("reading config: %v", err)
	}

	if err := yaml.Unmarshal(raw, &partialCfg); err != nil {
		tb.Fatalf("parsing config: %v", err)
	}

	networkCfg, err := partialCfg.Networks.Get(network)
	if err != nil {
		tb.Fatalf("getting network config: %v", err)
	}

	return NewApi(
		config.DataSource{URL: nodeURL},
		WithRateLimit(100),
		WithTimeout(60*time.Second),
		WithTraceMethod(networkCfg.GetTraceMethod()),
	)
}

func BenchmarkBlockBulkSize(b *testing.B) {
	api := newBenchAPI(b)
	ctx := context.Background()

	head, err := api.Head(ctx)
	if err != nil {
		b.Fatalf("getting head: %v", err)
	}

	startLevel := head - 100

	for bulkSize := 1; bulkSize <= requestBulkSize; bulkSize++ {
		levels := make([]pkgTypes.Level, bulkSize)
		for j := range levels {
			levels[j] = startLevel + pkgTypes.Level(j)
		}

		b.Run(fmt.Sprintf("bulk_%02d", bulkSize), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := api.BlockBulk(ctx, levels...)
				if err != nil {
					b.Fatalf("BlockBulk(size=%d): %v", bulkSize, err)
				}
			}

			avgPerBlock := float64(b.Elapsed()) / float64(b.N) / float64(bulkSize)
			b.ReportMetric(avgPerBlock/float64(time.Millisecond), "ms/block")
		})
	}
}

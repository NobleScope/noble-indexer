package rpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	indexerConfig "github.com/NobleScope/noble-indexer/pkg/indexer/config"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	goLibConfig "github.com/dipdup-net/go-lib/config"
)

const (
	requestBulkSize = 30
	apiRateLimit    = 100
	configPath      = "../../../configs/dipdup.yml"
)

func newBenchAPI(tb testing.TB) API {
	tb.Helper()

	var cfg indexerConfig.Config
	if err := goLibConfig.ParseWithValidator(configPath, nil, &cfg); err != nil {
		tb.Fatalf("parsing config: %v", err)
	}

	ds, ok := cfg.DataSources["node_rpc"]
	if !ok || ds.URL == "" {
		tb.Skip("node_rpc datasource is not configured")
	}

	networkCfg, err := cfg.Networks.Get(cfg.Network)
	if err != nil {
		tb.Fatalf("getting network config: %v", err)
	}

	return NewApi(
		ds,
		WithRateLimit(apiRateLimit),
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

			avgPerBulk := float64(b.Elapsed()) / float64(b.N)
			avgPerBlock := avgPerBulk / float64(bulkSize)
			b.ReportMetric(avgPerBulk/float64(time.Millisecond), "ms/bulk")
			b.ReportMetric(avgPerBlock/float64(time.Millisecond), "ms/block")
		})
	}
}

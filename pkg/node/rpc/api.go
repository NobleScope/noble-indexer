package rpc

import (
	"net/url"
	"time"

	"github.com/NobleScope/noble-indexer/pkg/node/rpc/trace_provider"
	"github.com/dipdup-net/go-lib/config"
	jsoniter "github.com/json-iterator/go"
	fastshot "github.com/opus-domini/fast-shot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

const (
	userAgent = "Noble Indexer"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type API struct {
	client        fastshot.ClientHttpMethods
	cfg           config.DataSource
	rateLimit     *rate.Limiter
	timeout       time.Duration
	traceProvider trace_provider.ITraceProvider
	log           zerolog.Logger
}

func NewApi(cfg config.DataSource, opts ...APIOption) API {
	nodeURL, err := url.Parse(cfg.URL)
	if err != nil {
		panic(err)
	}

	api := API{
		cfg:           cfg,
		client:        fastshot.NewClient(nodeURL.Scheme + "://" + nodeURL.Host).Build(),
		rateLimit:     rate.NewLimiter(rate.Every(time.Second/time.Duration(10)), 10),
		timeout:       time.Second * 30,
		traceProvider: &trace_provider.ParityTraceProvider{},
		log:           log.With().Str("module", "node rpc").Logger(),
	}

	for i := range opts {
		opts[i](&api)
	}

	return api
}

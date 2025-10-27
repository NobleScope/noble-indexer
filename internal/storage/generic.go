package storage

import (
	"context"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

var Models = []any{
	&Block{},
	&Tx{},
	&Log{},
	&Address{},
	&Contract{},
	&Trace{},
	&Balance{},
	&State{},
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type Transaction interface {
	sdk.Transaction

	SaveTransactions(ctx context.Context, txs ...*Tx) error
	SaveLogs(ctx context.Context, logs ...Log) error
	SaveTraces(ctx context.Context, traces ...*Trace) error
	SaveAddresses(ctx context.Context, addresses ...*Address) (int64, error)
	SaveBalances(ctx context.Context, balances ...*Balance) error
	SaveContracts(ctx context.Context, addresses ...*Contract) error

	State(ctx context.Context, name string) (state State, err error)
}

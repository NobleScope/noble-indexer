package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

var Models = []any{
	&Block{},
	&BlockStats{},
	&Tx{},
	&Transfer{},
	&Token{},
	&TokenBalance{},
	&Log{},
	&Address{},
	&Contract{},
	&Source{},
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
	SaveTransfers(ctx context.Context, transfers ...*Transfer) error
	SaveTokens(ctx context.Context, tokens ...*Token) error
	SaveTokenBalances(ctx context.Context, tokenBalances ...*TokenBalance) (tb []TokenBalance, err error)
	SaveTokenMetadata(ctx context.Context, tokens ...*Token) error
	SaveSources(ctx context.Context, sources ...*Source) error

	RollbackBlock(ctx context.Context, height types.Level) error
	RollbackBlockStats(ctx context.Context, height types.Level) (stats BlockStats, err error)
	RollbackAddresses(ctx context.Context, height types.Level) (addresses []Address, err error)
	RollbackTxs(ctx context.Context, height types.Level) (txs []Tx, err error)
	RollbackTraces(ctx context.Context, height types.Level) (traces []Trace, err error)
	RollbackLogs(ctx context.Context, height types.Level) error
	RollbackTransfers(ctx context.Context, height types.Level) (transfers []Transfer, err error)
	RollbackTokens(ctx context.Context, height types.Level) (tokens []Token, err error)
	RollbackContracts(ctx context.Context, height types.Level) error
	DeleteBalances(ctx context.Context, ids []uint64) error
	DeleteTokenBalances(ctx context.Context, tokenIds []string, contractIds []uint64, zeroBalances []*TokenBalance) error

	State(ctx context.Context, name string) (state State, err error)
	LastBlock(ctx context.Context) (block Block, err error)
}

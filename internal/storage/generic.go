package storage

import (
	"context"
	"io"

	"github.com/baking-bad/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/lib/pq"
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
	&ProxyContract{},
	&Source{},
	&Trace{},
	&Balance{},
	&State{},
	&ERC4337UserOp{},
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type Transaction interface {
	sdk.Transaction

	SaveTransactions(ctx context.Context, txs ...*Tx) error
	SaveLogs(ctx context.Context, logs ...*Log) error
	SaveTraces(ctx context.Context, traces ...*Trace) error
	SaveAddresses(ctx context.Context, addresses ...*Address) (int64, error)
	SaveBalances(ctx context.Context, balances ...*Balance) error
	SaveContracts(ctx context.Context, addresses ...*Contract) (int64, error)
	SaveTransfers(ctx context.Context, transfers ...*Transfer) error
	SaveTokens(ctx context.Context, tokens ...*Token) (int64, error)
	SaveTokenBalances(ctx context.Context, tokenBalances ...*TokenBalance) (tb []TokenBalance, err error)
	SaveTokenMetadata(ctx context.Context, tokens ...*Token) error
	SaveSources(ctx context.Context, sources ...*Source) error
	SaveProxyContracts(ctx context.Context, contracts ...*ProxyContract) error
	SaveERC4337UserOps(ctx context.Context, userOps ...*ERC4337UserOp) error

	RollbackBlock(ctx context.Context, height types.Level) error
	RollbackBlockStats(ctx context.Context, height types.Level) (stats BlockStats, err error)
	RollbackAddresses(ctx context.Context, height types.Level) (addresses []Address, err error)
	RollbackTxs(ctx context.Context, height types.Level) (txs []Tx, err error)
	RollbackTraces(ctx context.Context, height types.Level) (traces []Trace, err error)
	RollbackLogs(ctx context.Context, height types.Level) error
	RollbackTransfers(ctx context.Context, height types.Level) (transfers []Transfer, err error)
	RollbackTokens(ctx context.Context, height types.Level) (tokens []Token, err error)
	RollbackContracts(ctx context.Context, height types.Level) error
	RollbackERC4337UserOps(ctx context.Context, height types.Level) error
	DeleteBalances(ctx context.Context, ids []uint64) error
	DeleteTokenBalances(ctx context.Context, tokenIds []string, contractIds []uint64, zeroBalances []*TokenBalance) error

	State(ctx context.Context, name string) (state State, err error)
	LastBlock(ctx context.Context) (block Block, err error)
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type Notificator interface {
	Notify(ctx context.Context, channel string, payload string) error
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type Listener interface {
	io.Closer

	Subscribe(ctx context.Context, channels ...string) error
	Listen() chan *pq.Notification
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ListenerFactory interface {
	CreateListener() Listener
}

const (
	ChannelHead  = "head"
	ChannelBlock = "block"
)

type SearchResult struct {
	Id    uint64 `bun:"id"`
	Value string `bun:"value"`
	Type  string `bun:"type"`
}

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type ISearch interface {
	Search(ctx context.Context, query []byte, limit, offset int) ([]SearchResult, error)
	SearchText(ctx context.Context, text string, limit, offset int) ([]SearchResult, error)
}

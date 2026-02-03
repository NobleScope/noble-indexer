package context

import (
	"github.com/baking-bad/noble-indexer/internal/pool"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/sync"
)

type Context struct {
	Addresses      *sync.Map[string, *storage.Address]
	Contracts      *sync.Map[string, *storage.Contract]
	Tokens         *sync.Map[string, *storage.Token]
	TokenBalances  *sync.Map[string, *storage.TokenBalance]
	ProxyContracts *sync.Map[string, *storage.ProxyContract]
	ERC4337UserOps *sync.Map[string, *storage.ERC4337UserOp]
	Traces         *sync.Map[string, []*storage.Trace]

	Block *storage.Block
}

func NewContext() *Context {
	return &Context{
		Addresses:      sync.NewMap[string, *storage.Address](),
		Contracts:      sync.NewMap[string, *storage.Contract](),
		Tokens:         sync.NewMap[string, *storage.Token](),
		TokenBalances:  sync.NewMap[string, *storage.TokenBalance](),
		ProxyContracts: sync.NewMap[string, *storage.ProxyContract](),
		ERC4337UserOps: sync.NewMap[string, *storage.ERC4337UserOp](),
		Traces:         sync.NewMap[string, []*storage.Trace](),
	}
}

func (ctx *Context) AddAddress(address *storage.Address) {
	if address == nil {
		return
	}
	if addr, ok := ctx.Addresses.Get(address.String()); ok {
		addr.Interactions += address.Interactions
		addr.TxsCount += address.TxsCount
		addr.ContractsCount += address.ContractsCount
		address.Balance = addr.Balance
		if address.IsContract {
			addr.IsContract = true
		}

	} else {
		ctx.Addresses.Set(address.String(), address)
	}
}

func (ctx *Context) AddContract(contract *storage.Contract) {
	if contract == nil {
		return
	}
	if _, ok := ctx.Contracts.Get(contract.String()); !ok {
		ctx.Contracts.Set(contract.String(), contract)
	}
}

func (ctx *Context) AddTrace(trace *storage.Trace) {
	if trace == nil {
		return
	}

	if traces, ok := ctx.Traces.Get(trace.Tx.Hash.String()); ok {
		traces = append(traces, trace)
		ctx.Traces.Set(trace.Tx.Hash.String(), traces)
	} else {
		ctx.Traces.Set(trace.Tx.Hash.String(), []*storage.Trace{trace})
	}
}

func (ctx *Context) AddToken(token *storage.Token) {
	if token == nil {
		return
	}
	if _, ok := ctx.Tokens.Get(token.String()); !ok {
		ctx.Tokens.Set(token.String(), token)
	}
}

func (ctx *Context) AddTokenBalance(tokenBalance *storage.TokenBalance) {
	if tokenBalance == nil {
		return
	}
	if tb, ok := ctx.TokenBalances.Get(tokenBalance.String()); !ok {
		ctx.TokenBalances.Set(tokenBalance.String(), tokenBalance)
	} else {
		tb.Balance = tb.Balance.Add(tokenBalance.Balance)
	}
}

func (ctx *Context) AddProxyContract(proxyContract *storage.ProxyContract) {
	if proxyContract == nil {
		return
	}
	if _, ok := ctx.ProxyContracts.Get(proxyContract.String()); !ok {
		ctx.ProxyContracts.Set(proxyContract.String(), proxyContract)
	}
}

func (ctx *Context) AddUserOp(userOp *storage.ERC4337UserOp) {
	if userOp == nil {
		return
	}
	if _, ok := ctx.ERC4337UserOps.Get(userOp.String()); !ok {
		ctx.ERC4337UserOps.Set(userOp.String(), userOp)
	}
}

func (ctx *Context) GetAddresses() []*storage.Address {
	return ctx.Addresses.Values()
}

func (ctx *Context) GetContracts() []*storage.Contract {
	return ctx.Contracts.Values()
}

var tracesPool = pool.New(func() []*storage.Trace {
	return make([]*storage.Trace, 0, 1024)
})

func (ctx *Context) GetTraces() []*storage.Trace {
	traces := tracesPool.Get()
	defer func() {
		for i := range traces {
			traces[i] = nil
		}
		traces = traces[:0]
		tracesPool.Put(traces)
	}()

	for _, ts := range ctx.Traces.Values() {
		traces = append(traces, ts...)
	}
	return traces
}

func (ctx *Context) GetTracesByTxHash(txHash types.Hex) []*storage.Trace {
	if traces, ok := ctx.Traces.Get(txHash.String()); ok {
		return traces
	}
	return nil
}

func (ctx *Context) GetTokens() []*storage.Token {
	return ctx.Tokens.Values()
}

func (ctx *Context) GetTokenBalances() []*storage.TokenBalance {
	return ctx.TokenBalances.Values()
}

func (ctx *Context) GetProxyContracts() []*storage.ProxyContract {
	return ctx.ProxyContracts.Values()
}

func (ctx *Context) GetUserOps() []*storage.ERC4337UserOp {
	return ctx.ERC4337UserOps.Values()
}

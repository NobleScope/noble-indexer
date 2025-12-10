package context

import (
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

func (ctx *Context) GetAddresses() []*storage.Address {
	addresses := make([]*storage.Address, 0)
	addresses = append(addresses, ctx.Addresses.Values()...)

	return addresses
}

func (ctx *Context) GetContracts() []*storage.Contract {
	contracts := make([]*storage.Contract, 0)
	contracts = append(contracts, ctx.Contracts.Values()...)

	return contracts
}

func (ctx *Context) GetTraces() []*storage.Trace {
	traces := make([]*storage.Trace, 0)
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
	tokens := make([]*storage.Token, 0)
	tokens = append(tokens, ctx.Tokens.Values()...)

	return tokens
}

func (ctx *Context) GetTokenBalances() []*storage.TokenBalance {
	tokenBalances := make([]*storage.TokenBalance, 0)
	tokenBalances = append(tokenBalances, ctx.TokenBalances.Values()...)

	return tokenBalances
}

func (ctx *Context) GetProxyContracts() []*storage.ProxyContract {
	return ctx.ProxyContracts.Values()
}

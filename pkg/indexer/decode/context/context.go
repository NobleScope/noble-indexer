package context

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/indexer-sdk/pkg/sync"
)

type Context struct {
	Addresses *sync.Map[string, *storage.Address]
	Contracts *sync.Map[string, *storage.Contract]
	Traces    *sync.Map[string, []*storage.Trace]

	Block *storage.Block
}

func NewContext() *Context {
	return &Context{
		Addresses: sync.NewMap[string, *storage.Address](),
		Contracts: sync.NewMap[string, *storage.Contract](),
		Traces:    sync.NewMap[string, []*storage.Trace](),
	}
}

func (ctx *Context) AddAddress(address *storage.Address) {
	if address == nil {
		return
	}
	if _, ok := ctx.Addresses.Get(address.String()); !ok {
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

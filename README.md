# noble-indexer
Noble indexer

[![Swagger](https://img.shields.io/badge/API-Swagger-green)](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/NobleScope/noble-indexer/master/cmd/api/docs/swagger.json)

## API Documentation

Interactive API documentation is available via Swagger UI:
- [Swagger UI](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/NobleScope/noble-indexer/master/cmd/api/docs/swagger.json) - interactive API explorer
- Local: http://localhost:9876/swagger/index.html (when running the API server)
- Swagger JSON: [cmd/api/docs/swagger.json](https://github.com/NobleScope/noble-indexer/blob/master/cmd/api/docs/swagger.json)
- Swagger YAML: [cmd/api/docs/swagger.yaml](https://github.com/NobleScope/noble-indexer/blob/master/cmd/api/docs/swagger.yaml)

To regenerate API documentation:
```bash
make api-docs
```

## Node Requirements

This indexer requires an Ethereum execution client with specific RPC methods enabled. Below are the required endpoints, rate limiting considerations, and client-specific configurations.

### Required RPC Methods

| Method | Description | Usage |
|--------|-------------|-------|
| `eth_blockNumber` | Get latest block number | Sync status checking |
| `eth_getBlockByNumber` | Fetch block data by number | Block indexing (batched) |
| `eth_getBlockReceipts` | Get transaction receipts for a block | Receipt indexing (batched) |
| `eth_getStorageAt` | Read contract storage slots | Proxy contract resolution |
| `eth_call` | Execute contract calls | Token metadata (name, symbol, decimals, URI) |
| `trace_block` | Get execution traces for a block (Parity/Erigon style) | Internal transaction indexing (batched) |
| `debug_traceBlockByNumber` | Get execution traces for a block (Geth style) | Internal transaction indexing (batched)|

### WebSocket Requirements

| Subscription | Description |
|--------------|-------------|
| `eth_subscribe` (newHeads) | Real-time new block header notifications |

The indexer uses WebSocket connection for real-time block synchronization. Configure `node_ws` data source with your WebSocket endpoint.

### Rate Limiting

- **Default rate limit**: 10 requests per second
- **Configurable**: Set via `RequestsPerSecond` in data source configuration
- **Batch requests**: The indexer uses batch JSON-RPC calls (3 requests per block level) to optimize throughput

### Timeout Configuration

- **Default timeout**: 30 seconds
- **Configurable**: Set via `Timeout` in data source configuration

### Client-Specific Configuration

#### Geth (go-ethereum)

Geth does not support `trace_block` natively. You need to use `debug_traceBlockByNumber` or run Geth with a custom tracer.

**Option 1**: Use a fork with trace support (recommended)
- Use [geth with Parity-style tracing](https://github.com/ledgerwatch/erigon) or switch to Erigon/Reth

**Option 2**: Enable debug API
```bash
geth --http --http.api eth,net,web3,debug --ws --ws.api eth,net,web3,debug
```

> **Note**: Standard Geth `debug_traceBlockByNumber` has a different response format. The indexer expects Parity/OpenEthereum-style `trace_block` responses.

#### Reth

Reth has native support for all required methods including `trace_block`.

```bash
reth node \
  --http \
  --http.api eth,net,web3,trace \
  --ws \
  --ws.api eth,net,web3,trace
```

#### Erigon

Erigon supports all required methods including `trace_block` natively.

```bash
erigon \
  --http \
  --http.api eth,erigon,trace,web3,net \
  --ws
```

#### Nethermind

Nethermind supports `trace_block` via the Trace module.

```bash
nethermind \
  --JsonRpc.Enabled true \
  --JsonRpc.EnabledModules "Eth,Net,Web3,Trace" \
  --Init.WebSocketsEnabled true
```

#### Besu

Hyperledger Besu supports tracing via the `trace` API.

```bash
besu \
  --rpc-http-enabled \
  --rpc-http-api=ETH,NET,WEB3,TRACE \
  --rpc-ws-enabled \
  --rpc-ws-api=ETH,NET,WEB3,TRACE
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `NOBLE_NODE_URL` | HTTP RPC endpoint URL | `http://localhost:8545` |
| `INDEXER_REQUEST_BULK_SIZE` | Number of blocks per batch request | `3` |

### Recommended Node Setup

For optimal indexer performance:

1. **Archive node**: Required for `eth_getStorageAt` on historical blocks
2. **Tracing enabled**: Required for `trace_block` to index internal transactions
3. **WebSocket support**: Required for real-time block synchronization
4. **Sufficient rate limits**: Minimum 10 RPS recommended, higher for faster sync
5. **Low latency**: Local node or dedicated RPC endpoint recommended

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


## Running for a New Network

### 1. Add Genesis Block

Place the genesis block JSON file into the [`assets/`](https://github.com/NobleScope/noble-indexer/tree/master/assets) directory:

```
assets/genesis-<network-name>.json
```

Then set the `GENESIS_FILENAME` env variable to match the file name (e.g. `genesis-<network-name>.json`).

### 2. Register the Network

Add a new entry in the `networks` section of [`configs/dipdup.yml`](https://github.com/NobleScope/noble-indexer/blob/master/configs/dipdup.yml):

```yaml
networks:
  my-network:
    precompiled_contracts:
      - 0x6625300000000000000000000000000000000000
    trace_method: trace_block  # or debug_traceBlockByNumber for Geth-style nodes
```

- `precompiled_contracts` — list of precompiled contract addresses for the network (can be empty).
- `trace_method` — `trace_block` (Parity/Erigon/Reth) or `debug_traceBlockByNumber` (Geth).

Set the `NETWORK` env variable to the name of the added network (e.g. `my-network`).

### 3. Configure Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```bash
cp .env.example .env
```

| Variable | Required | Description |
|----------|----------|-------------|
| `INDEXER_NAME` | yes | Unique indexer instance name |
| `INDEXER_REQUEST_BULK_SIZE` | yes | Number of blocks fetched per batch (e.g. `15`) |
| `INDEXER_SCRIPTS_DIR` | yes | Path to SQL scripts directory (`./database`) |
| `INDEXER_START_LEVEL` | yes | Block height to start indexing from (`0` for genesis) |
| `EVM_NODE_RPS` | yes | Max requests per second to the node |
| `EVM_NODE_URL` | yes | HTTP RPC endpoint (e.g. `https://ethereum-rpc.publicnode.com`) |
| `EVM_NODE_WS_URL` | no | WebSocket endpoint. **Optional** — omit this variable if you don't need real-time block subscriptions via WebSocket |
| `POSTGRES_DB` | yes | PostgreSQL database name |
| `POSTGRES_HOST` | yes | PostgreSQL host |
| `POSTGRES_USER` | yes | PostgreSQL user |
| `POSTGRES_PASSWORD` | yes | PostgreSQL password |
| `LOG_LEVEL` | yes | Logging level (`info`, `debug`, `warn`, `error`) |
| `NETWORK` | yes | Network name matching the entry in `dipdup.yml` |
| `GENESIS_FILENAME` | yes | Genesis file name in `assets/` |
| `CACHE_URL` | yes | Valkey/Redis URL (e.g. `redis://cache:6379`) |
| `TOKEN_RESOLVER_NAME` | yes | Token metadata resolver instance name |
| `TOKEN_RESOLVER_REQUEST_BULK_SIZE` | yes | Batch size for token metadata resolution |
| `TOKEN_RESOLVER_SYNC_PERIOD` | yes | Sync period in seconds |
| `CONTRACT_RESOLVER_NAME` | yes | Contract metadata resolver instance name |
| `CONTRACT_RESOLVER_SYNC_PERIOD` | yes | Sync period in seconds |
| `PROXY_NODE_BATCH_SIZE` | yes | Batch size for proxy contract resolution |

### 4. Start Valkey (Cache)

Valkey is required for caching token metadata. Start it via Docker:

```bash
docker compose up -d cache
```

This will start a Valkey instance on port `6379`. The `CACHE_URL` env variable should point to it (e.g. `redis://cache:6379` when using docker-compose, or `redis://localhost:6379` when running the indexer outside of Docker).

### 5. Start the Services

```bash
# Start everything (database, indexer, API, resolvers, cache)
docker compose up -d

# Or start only the indexer + API
docker compose up -d db cache indexer api
```

To run locally without Docker:

```bash
# Start dependencies
docker compose up -d db cache

# Run the indexer
make indexer

# Run the API (in a separate terminal)
make api
```

### Recommended Node Setup

For optimal indexer performance:

1. **Archive node**: Required for `eth_getStorageAt` on historical blocks
2. **Tracing enabled**: Required for `trace_block` to index internal transactions
3. **WebSocket support**: Required for real-time block synchronization
4. **Sufficient rate limits**: Minimum 10 RPS recommended, higher for faster sync
5. **Low latency**: Local node or dedicated RPC endpoint recommended

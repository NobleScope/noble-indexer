# Noble Indexer — CLAUDE.md

## Project Overview

Go-based blockchain indexer + REST API for the Noble blockchain (Ethereum-compatible). Indexes blocks, transactions, traces, logs, contracts, and tokens into PostgreSQL (TimescaleDB) and exposes them via Echo HTTP/WebSocket API.

## Architecture

```
cmd/
  indexer/                     # Blockchain indexer daemon
  api/main.go                  # REST API server
    handler/                   # Echo handlers (one file per entity)
  contract_metadata_resolver/  # IPFS contract metadata worker
  token_metadata_resolver/     # IPFS token metadata worker
  contract_verifier/           # Contract source verification worker
  common/                      # Shared init: config, DB, cache, logger, profiler
pkg/
  indexer/                     # Core indexer pipeline
    receiver/                  # Fetches blocks from node RPC/WS
    parser/                    # Decodes raw block data, loads ABIs from assets/
    storage/                   # Saves parsed data to DB in one DB transaction
    rollback/                  # Handles blockchain reorganizations
    genesis/                   # Handles genesis block separately
    proxy_contracts_resolver/  # Resolves proxy contract implementations
    decode/context/            # Context object passed between parser → storage
  types/                       # pkg-level domain types: Block, Tx, Trace, Log, Hex, Level
  node/rpc/                    # Ethereum JSON-RPC + WebSocket client
internal/
  storage/postgres/            # Bun ORM models + all DB query implementations
    scopes.go                  # Reusable query filters and pagination helpers
    transaction.go             # DB transaction: save/rollback all entities
    core.go                    # DB init, migrations, hypertables, enums, indexes
  cache/                       # Valkey/Redis caching layer
  pool/                        # sync.Pool wrappers for reusing slices
database/
  functions/ views/            # SQL applied at startup via INDEXER_SCRIPTS_DIR
configs/
  dipdup.yml                   # YAML config with ${ENV_VAR} substitution
```

**Indexer pipeline:** Node RPC/WS → Receiver → Parser → Storage module → PostgreSQL

**Module wiring** (in `pkg/indexer/indexer.go`): modules connect via named inputs/outputs using `module.AttachTo(source, outputName, inputName)`. Every module has a `StopOutput` that feeds into the stopper.

## Key Libraries

| Purpose | Library |
|---------|---------|
| HTTP | `github.com/labstack/echo/v4` |
| ORM | `github.com/uptrace/bun` + `lib/pq` |
| Ethereum | `github.com/ethereum/go-ethereum` |
| Cache | `github.com/valkey-io/valkey-go` |
| Logging | `github.com/rs/zerolog` |
| Validation | `github.com/go-playground/validator` |
| Errors | `github.com/pkg/errors` |
| Mocks | `go.uber.org/mock/mockgen` |
| Swagger | `github.com/swaggo/swag` |
| Indexer SDK | `github.com/dipdup-net/indexer-sdk` |

## Commands

```bash
make indexer    # go run ./cmd/indexer -c ./configs/dipdup.yml
make api        # go run ./cmd/api -c ./configs/dipdup.yml
make api-docs   # swag init (regenerate Swagger)
make test       # go test -p 8 -timeout 120s ./...
make generate   # go generate ./...  (regenerate mocks)
make lint       # golangci-lint
```

## Configuration

YAML config with `${ENV_VAR}` substitution (`configs/dipdup.yml`):

```
NOBLE_NODE_URL / NOBLE_NODE_WS_URL / NOBLE_NODE_RPS
POSTGRES_HOST / PORT / USER / PASSWORD / DB / MAX_OPEN_CONNECTIONS
API_HOST / API_PORT / API_RATE_LIMIT / API_REQUEST_TIMEOUT / API_WEBSOCKET_ENABLED
CACHE_URL / CACHE_TTL
INDEXER_START_LEVEL / INDEXER_SCRIPTS_DIR / NETWORK
```

## Storage Patterns

All storage files in `internal/storage/postgres/`. Each entity has its own file (`tx.go`, `block.go`, etc.).

**Typical query pattern** — subquery for filters, outer query for JOINs:
```go
func (t *Tx) Filter(ctx context.Context, filter storage.TxListFilter) ([]storage.Tx, error) {
    query := t.DB().NewSelect().Model(&txs)
    query = txListFilter(query, filter)          // apply filters from scopes.go

    outerQuery := t.DB().NewSelect().
        ColumnExpr("tx.*").
        ColumnExpr("from_addr.hash AS from_address__hash, ...").  // flat JOIN expansion
        TableExpr("(?) AS tx", query).
        Join("LEFT JOIN address AS from_addr ON from_addr.id = tx.from_address_id")

    return txs, outerQuery.Scan(ctx, &txs)
}
```

**Joined relation columns** use `__` separator: `"from_addr.hash AS from_address__hash"` maps to `Tx.FromAddress.Hash`.

**Pagination helpers** (`scopes.go`):
- `limitScope(q, limit)` — clamps 1–100, default 10
- `sortScope(q, field, order)` — single field sort
- `sortTimeIDScope(q, order)` — always sort by `time, id` (used for time-series tables)
- `sortMultipleScope(q, []SortField{...})` — multi-field sort
- All `*ListFilter` functions take their filter struct and return the modified query

**DB transaction** for saving a block (`transaction.go`):
```go
tx, _ := postgres.BeginTransaction(ctx, module.storage)
defer tx.Close(ctx)
// tx.Add(), tx.Update(), tx.Flush() — then tx.HandleError() on failure
```

## API Handler Pattern

```go
// 1. Request struct with Echo binding tags + validator tags
type blockListRequest struct {
    Limit  int    `query:"limit"  validate:"omitempty,min=1,max=100"`
    Offset int    `query:"offset" validate:"omitempty,min=0"`
    Sort   string `query:"sort"   validate:"omitempty,oneof=asc desc"`
}
func (r *blockListRequest) SetDefault() { r.Limit = 10; r.Sort = "asc" }

// 2. Swagger annotations above every handler
// @Summary   Get block
// @Tags      block
// @Param     height path integer true "Block height" minimum(1)
// @Success   200 {object} responses.Block
// @Router    /blocks/{height} [get]
func (h *BlockHandler) Get(c echo.Context) error {
    req, err := bindAndValidate[blockListRequest](c)  // generic helper in requests.go
    if err != nil { return badRequestError(c, err) }
    req.SetDefault()

    result, err := h.block.Filter(c.Request().Context(), storage.BlockListFilter{...})
    if err != nil { return handleError(c, err, h.block) }

    return returnArray(c, response)   // returns [] not null for empty
}
```

**Helper functions** (`requests.go`, `error.go`, `constant.go`):
- `bindAndValidate[T](c)` — generic bind + validate
- `badRequestError(c, err)` / `handleError(c, err, storage)` — consistent error responses
- `returnArray(c, arr)` — returns empty array `[]` instead of `null`
- `pgSort(s string) sdk.SortOrder` — converts "asc"/"desc" string to SDK type
- `StringArray` — comma-separated query param `?types=erc20,erc721`

## Indexer Module Pattern

Each pipeline module embeds `modules.BaseModule`, has named string constants for inputs/outputs, and follows this structure:

```go
const (
    InputName  = "blocks"
    OutputName = "data"
    StopOutput = "stop"
)

type Module struct {
    modules.BaseModule
    // dependencies...
}

func NewModule(...) Module {
    m := Module{BaseModule: modules.New("parser"), ...}
    m.CreateInputWithCapacity(InputName, 128)
    m.CreateOutput(OutputName)
    m.CreateOutput(StopOutput)
    return m
}

func (m *Module) Start(ctx context.Context) {
    m.G.GoCtx(ctx, m.listen)
}

func (m *Module) listen(ctx context.Context) {
    input := m.MustInput(InputName)
    for {
        select {
        case <-ctx.Done(): return
        case msg, ok := <-input.Listen():
            if !ok { m.MustOutput(StopOutput).Push(struct{}{}); continue }
            // process msg...
            m.MustOutput(OutputName).Push(result)
        }
    }
}

func (m *Module) Close() error { m.G.Wait(); return nil }
```

**ABI files** are loaded from disk at startup by the parser: `{AssetsDir}/abi/erc20.json`, `erc721.json`, `erc1155.json`, `ERC4337/*.json`.

## Adding a New Entity (Checklist)

1. `internal/storage/` — add model struct + filter struct + interface `IFoo`
2. `internal/storage/postgres/foo.go` — implement queries using subquery+JOIN pattern
3. `internal/storage/postgres/core.go` — register in `Storage` struct, create hypertable if time-series
4. `internal/storage/postgres/index.go` — add indexes
5. `internal/storage/postgres/transaction.go` — add `SaveFoo` + `RollbackFoo`
6. Mock: add `//go:generate` directive, run `make generate`
7. Parser: add parsing logic, add to `decode/context/`
8. `pkg/indexer/storage/storage.go` — call `saveFoo` in `processBlockInTransaction`
9. `cmd/api/handler/foo.go` — handler with Swagger annotations
10. Register routes in `cmd/api/main.go`
11. Run `make api-docs`

## Key Conventions

- `zerolog` only for logging — never `fmt.Print` in production paths
- `errors.Wrap(err, "context")` from `github.com/pkg/errors`
- Storage interfaces only — don't use concrete postgres types outside `internal/`
- WebSocket notifications are skipped during initial sync (`time.Since(block.Time) > time.Hour`)
- `pool.New(func() []T)` — use `internal/pool` for reusing slices in hot paths
- Active linters to watch: `zerologlint`, `musttag`, `gosec`, `containedctx`

## Testing

- Mocks are auto-generated — never edit manually
- `testfixtures` for DB integration tests
- Run `make test` before committing

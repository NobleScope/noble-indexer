package postgres

import (
	"context"
	"errors"

	models "github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

type Transaction struct {
	storage.Transaction
}

func BeginTransaction(ctx context.Context, tx storage.Transactable) (models.Transaction, error) {
	t, err := tx.BeginTransaction(ctx)
	return Transaction{t}, err
}

func (tx Transaction) SaveTransactions(ctx context.Context, txs ...*models.Tx) error {
	switch len(txs) {
	case 0:
		return nil
	case 1:
		return tx.Add(ctx, txs[0])
	default:
		arr := make([]any, len(txs))
		for i := range txs {
			arr[i] = txs[i]
		}
		return tx.BulkSave(ctx, arr)
	}
}

func (tx Transaction) SaveLogs(ctx context.Context, logs ...*models.Log) error {
	switch len(logs) {
	case 0:
		return nil
	case 1:
		return tx.Add(ctx, logs[0])
	default:
		arr := make([]any, len(logs))
		for i := range logs {
			arr[i] = logs[i]
		}
		return tx.BulkSave(ctx, arr)
	}
}

func (tx Transaction) SaveSources(ctx context.Context, sources ...*models.Source) error {
	switch len(sources) {
	case 0:
		return nil
	case 1:
		return tx.Add(ctx, sources[0])
	default:
		arr := make([]any, len(sources))
		for i := range sources {
			arr[i] = sources[i]
		}
		return tx.BulkSave(ctx, arr)
	}
}

type addedAddress struct {
	bun.BaseModel `bun:"address"`
	*models.Address

	Xmax uint64 `bun:"xmax"`
}

func (tx Transaction) SaveAddresses(ctx context.Context, addresses ...*models.Address) (int64, error) {
	if len(addresses) == 0 {
		return 0, nil
	}

	addr := make([]addedAddress, len(addresses))
	for i := range addresses {
		addr[i].Address = addresses[i]
	}

	_, err := tx.Tx().NewInsert().Model(&addr).
		Column("hash", "first_height", "last_height", "is_contract", "txs_count", "contracts_count", "interactions").
		On("CONFLICT (hash) DO UPDATE").
		Set("last_height = GREATEST(EXCLUDED.last_height, added_address.last_height)").
		Set("txs_count = EXCLUDED.txs_count + added_address.txs_count").
		Set("contracts_count = EXCLUDED.contracts_count + added_address.contracts_count").
		Set("interactions = EXCLUDED.interactions + added_address.interactions").
		Returning("xmax, id").
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	for i := range addr {
		if addr[i].Xmax == 0 {
			count++
		}
	}

	return count, err
}

func (tx Transaction) SaveBalances(ctx context.Context, balances ...*models.Balance) error {
	if len(balances) == 0 {
		return nil
	}

	_, err := tx.Tx().NewInsert().Model(&balances).
		Column("id", "value").
		On("CONFLICT (id) DO UPDATE").
		Set("value = EXCLUDED.value + balance.value").
		Exec(ctx)

	return err
}

func (tx Transaction) SaveContracts(ctx context.Context, contracts ...*models.Contract) error {
	if len(contracts) == 0 {
		return nil
	}

	_, err := tx.Tx().NewInsert().Model(&contracts).
		On("CONFLICT (id) DO UPDATE").
		Set("verified = EXCLUDED.verified").
		Set("abi = EXCLUDED.abi").
		Set("compiler_version = EXCLUDED.compiler_version").
		Set("metadata_link = EXCLUDED.metadata_link").
		Set("language = EXCLUDED.language").
		Set("optimizer_enabled = EXCLUDED.optimizer_enabled").
		Set("tags = EXCLUDED.tags").
		Set("status = EXCLUDED.status").
		Set("retry_count = EXCLUDED.retry_count").
		Set("error = EXCLUDED.error").
		Set("updated_at = now()").
		Exec(ctx)

	return err
}

func (tx Transaction) SaveTraces(ctx context.Context, traces ...*models.Trace) error {
	switch len(traces) {
	case 0:
		return nil
	case 1:
		return tx.Add(ctx, traces[0])
	default:
		arr := make([]any, len(traces))
		for i := range traces {
			arr[i] = traces[i]
		}
		return tx.BulkSave(ctx, arr)
	}
}

func (tx Transaction) SaveTransfers(ctx context.Context, transfers ...*models.Transfer) error {
	switch len(transfers) {
	case 0:
		return nil
	case 1:
		return tx.Add(ctx, transfers[0])
	default:
		arr := make([]any, len(transfers))
		for i := range transfers {
			arr[i] = transfers[i]
		}
		return tx.BulkSave(ctx, arr)
	}
}

func (tx Transaction) SaveTokens(ctx context.Context, tokens ...*models.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	_, err := tx.Tx().NewInsert().Model(&tokens).
		On("CONFLICT (token_id, contract_id) DO UPDATE").
		Set("transfers_count = token.transfers_count + EXCLUDED.transfers_count").
		Set("supply = token.supply + EXCLUDED.supply").
		Set("last_height = EXCLUDED.last_height").
		Exec(ctx)

	return err
}

func (tx Transaction) SaveTokenMetadata(ctx context.Context, tokens ...*models.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	_, err := tx.Tx().NewInsert().Model(&tokens).
		On("CONFLICT (token_id, contract_id) DO UPDATE").
		Set("name = EXCLUDED.name").
		Set("symbol = EXCLUDED.symbol").
		Set("decimals = EXCLUDED.decimals").
		Set("status = EXCLUDED.status").
		Set("metadata_link = EXCLUDED.metadata_link").
		Set("metadata = EXCLUDED.metadata").
		Set("retry_count = EXCLUDED.retry_count").
		Set("error = EXCLUDED.error").
		Set("updated_at = now()").
		Exec(ctx)

	return err
}

func (tx Transaction) SaveTokenBalances(ctx context.Context, tokens ...*models.TokenBalance) ([]models.TokenBalance, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	var tbs []models.TokenBalance
	err := tx.Tx().NewInsert().Model(&tokens).
		On("CONFLICT (address_id, contract_id, token_id) DO UPDATE").
		Set("balance = token_balance.balance + EXCLUDED.balance").
		Returning("*").
		Scan(ctx, &tbs)

	if err != nil {
		return nil, err
	}

	return tbs, nil
}

func (tx Transaction) SaveProxyContracts(ctx context.Context, contracts ...*models.ProxyContract) error {
	if len(contracts) == 0 {
		return nil
	}

	_, err := tx.Tx().NewInsert().Model(&contracts).
		On("CONFLICT (id) DO UPDATE").
		Set(`implementation_id = CASE
				WHEN proxy_contract.implementation_id IS NULL OR EXCLUDED.height > proxy_contract.height
				THEN EXCLUDED.implementation_id
				ELSE proxy_contract.implementation_id
			END`).
		Set("height = GREATEST(EXCLUDED.height, proxy_contract.height)").
		Set(`status = CASE
				WHEN proxy_contract.implementation_id IS NULL OR EXCLUDED.height > proxy_contract.height
				THEN EXCLUDED.status
				ELSE proxy_contract.status
			END`).
		Set("resolving_attempts = EXCLUDED.resolving_attempts").
		Exec(ctx)

	return err
}

func (tx Transaction) RollbackBlock(ctx context.Context, height types.Level) error {
	_, err := tx.Tx().NewDelete().
		Model((*models.Block)(nil)).
		Where("height = ?", height).
		Exec(ctx)
	return err
}

func (tx Transaction) RollbackBlockStats(ctx context.Context, height types.Level) (stats models.BlockStats, err error) {
	_, err = tx.Tx().NewDelete().
		Model(&stats).
		Where("height = ?", height).
		Returning("*").
		Exec(ctx)
	return
}

func (tx Transaction) RollbackAddresses(ctx context.Context, height types.Level) (address []models.Address, err error) {
	_, err = tx.Tx().NewDelete().
		Model(&address).
		Where("height = ?", height).
		Returning("*").
		Exec(ctx)
	return
}

func (tx Transaction) RollbackTxs(ctx context.Context, height types.Level) (txs []models.Tx, err error) {
	_, err = tx.Tx().NewDelete().
		Model(&txs).
		Where("height = ?", height).
		Returning("*").
		Exec(ctx)
	return
}

func (tx Transaction) RollbackLogs(ctx context.Context, height types.Level) (err error) {
	_, err = tx.Tx().NewDelete().
		Model((*models.Log)(nil)).
		Where("height = ?", height).
		Exec(ctx)
	return
}

func (tx Transaction) RollbackTraces(ctx context.Context, height types.Level) (traces []models.Trace, err error) {
	_, err = tx.Tx().NewDelete().
		Model(&traces).
		Where("height = ?", height).
		Returning("*").
		Exec(ctx)
	return
}

func (tx Transaction) RollbackTransfers(ctx context.Context, height types.Level) (transfers []models.Transfer, err error) {
	_, err = tx.Tx().NewDelete().
		Model(&transfers).
		Where("height = ?", height).
		Returning("*").
		Exec(ctx)
	return
}

func (tx Transaction) RollbackTokens(ctx context.Context, height types.Level) (tokens []models.Token, err error) {
	_, err = tx.Tx().NewDelete().
		Model(&tokens).
		Where("height = ?", height).
		Returning("*").
		Exec(ctx)
	return
}

func (tx Transaction) RollbackContracts(ctx context.Context, height types.Level) (err error) {
	_, err = tx.Tx().NewDelete().
		Model((*models.Contract)(nil)).
		Where("height = ?", height).
		Exec(ctx)
	return
}

func (tx Transaction) DeleteBalances(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := tx.Tx().NewDelete().
		Model((*models.Balance)(nil)).
		Where("id IN (?)", bun.In(ids)).
		Exec(ctx)
	return err
}

func (tx Transaction) DeleteTokenBalances(ctx context.Context, tokenIds []string, contractIds []uint64, zeroBalances []*models.TokenBalance) error {
	if len(tokenIds) != len(contractIds) {
		return errors.New("tokenIds and contractIds must have same length")
	}

	query := tx.Tx().
		NewDelete().
		Model((*models.TokenBalance)(nil))

	query = query.WhereGroup("OR", func(q *bun.DeleteQuery) *bun.DeleteQuery {
		for i := range tokenIds {
			q = q.WhereOr(
				"(token_id = ?::numeric AND contract_id = ?)",
				tokenIds[i],
				contractIds[i],
			)
		}

		for _, t := range zeroBalances {
			q = q.WhereOr(
				"(token_id = ?::numeric AND contract_id = ? AND address_id = ?)",
				t.TokenID,
				t.ContractID,
				t.AddressID,
			)
		}

		return q
	})

	_, err := query.Exec(ctx)

	return err
}

func (tx Transaction) State(ctx context.Context, name string) (state models.State, err error) {
	err = tx.Tx().NewSelect().
		Model(&state).
		Where("name = ?", name).
		Scan(ctx)
	return
}

func (tx Transaction) LastBlock(ctx context.Context) (block models.Block, err error) {
	err = tx.Tx().NewSelect().
		Model(&block).
		Order("id DESC").
		Limit(1).
		Scan(ctx)
	return
}

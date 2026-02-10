package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// Contract -
type Contract struct {
	*postgres.Table[*storage.Contract]
}

// NewContract -
func NewContract(db *database.Bun) *Contract {
	return &Contract{
		Table: postgres.NewTable[*storage.Contract](db),
	}
}

// PendingMetadata - returns list of contracts with pending metadata resolution
func (c *Contract) PendingMetadata(ctx context.Context, retryDelay time.Duration, limit int) (contracts []*storage.Contract, err error) {
	threshold := time.Now().UTC().Add(-retryDelay)
	err = c.DB().NewSelect().
		Model(&contracts).
		Relation("Address").
		Where("metadata_link IS NOT NULL AND metadata_link <> ''").
		Where("status = 'pending' AND (updated_at < ? OR retry_count = 0)", threshold).
		Order("id ASC").
		Limit(limit).
		Scan(ctx)

	return
}

// ListWithTx - returns list of contracts with transaction and address info
func (c *Contract) ListWithTx(ctx context.Context, filters storage.ContractListFilter) (contracts []storage.Contract, err error) {
	query := c.DB().NewSelect().
		Model(&contracts)

	query = contractListFilter(query, filters)
	err = c.DB().NewSelect().TableExpr("(?) AS contract", query).
		ColumnExpr("contract.id, contract.height, contract.verified, contract.tx_id, contract.deployer_id, contract.compiler_version, contract.metadata_link, contract.language, contract.optimizer_enabled, contract.tags, contract.status, contract.retry_count, contract.error").
		ColumnExpr("address.hash AS address__hash").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("deployer.hash AS deployer__hash").
		Join("LEFT JOIN address ON address.id = contract.id").
		Join("LEFT JOIN tx ON contract.tx_id = tx.id").
		Join("LEFT JOIN address as deployer ON deployer.id = deployer_id").
		Scan(ctx, &contracts)

	return
}

// ByHash - returns contract by address hash
func (c *Contract) ByHash(ctx context.Context, hash pkgTypes.Hex) (contract storage.Contract, err error) {
	query := c.DB().NewSelect().
		Model((*storage.Address)(nil)).
		Where("hash = ?", hash)

	err = c.DB().NewSelect().
		TableExpr("(?) AS address", query).
		ColumnExpr("contract.id, contract.height, contract.verified, contract.tx_id, contract.deployer_id, contract.compiler_version, contract.metadata_link, contract.language, contract.optimizer_enabled, contract.tags, contract.status, contract.retry_count, contract.error").
		ColumnExpr("address.id AS address__id, address.first_height AS address__first_height, address.last_height AS address__last_height, address.hash AS address__hash, address.is_contract AS address__is_contract, address.txs_count AS address__txs_count, address.contracts_count AS address__contracts_count, address.interactions AS address__interactions").
		ColumnExpr("tx.hash AS tx__hash").
		ColumnExpr("implementation_address.hash AS implementation").
		ColumnExpr("deployer.hash AS deployer__hash").
		Join("JOIN contract ON contract.id = address.id").
		Join("LEFT JOIN proxy_contract ON proxy_contract.id = contract.id").
		Join("LEFT JOIN address AS implementation_address ON implementation_address.id = proxy_contract.implementation_id").
		Join("LEFT JOIN tx ON contract.tx_id = tx.id").
		Join("LEFT JOIN address as deployer ON deployer.id = deployer_id").
		Scan(ctx, &contract)

	return
}

// Code - returns contract code and ABI by contract address hash
func (c *Contract) Code(ctx context.Context, hash pkgTypes.Hex) (pkgTypes.Hex, json.RawMessage, error) {
	subquery := c.DB().NewSelect().
		Model((*storage.Address)(nil)).
		Where("hash = ?", hash).
		Column("id").
		Limit(1)

	var contract storage.Contract
	err := c.DB().NewSelect().
		Model(&contract).
		Where("id = (?)", subquery).
		Column("code", "abi").
		Scan(ctx)
	return contract.Code, contract.ABI, err
}

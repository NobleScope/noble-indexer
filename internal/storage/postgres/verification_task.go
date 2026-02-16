package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// VerificationTask -
type VerificationTask struct {
	*postgres.Table[*storage.VerificationTask]
}

// NewVerificationTask -
func NewVerificationTask(db *database.Bun) *VerificationTask {
	return &VerificationTask{
		Table: postgres.NewTable[*storage.VerificationTask](db),
	}
}

// Latest -
func (t *VerificationTask) Latest(ctx context.Context) (task storage.VerificationTask, err error) {
	err = t.DB().NewSelect().
		Model(&task).
		Where("status = 'VerificationStatusNew'").
		Order("creation_time ASC").
		Limit(1).
		Scan(ctx, &task)

	return
}

// ByContractId -
func (t *VerificationTask) ByContractId(ctx context.Context, contractId uint64) (tasks []storage.VerificationTask, err error) {
	err = t.DB().NewSelect().
		Model(&tasks).
		Where("contract_id = ?", contractId).
		Scan(ctx, &tasks)

	return
}

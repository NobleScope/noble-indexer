package postgres

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/database"
	"github.com/dipdup-net/indexer-sdk/pkg/storage/postgres"
)

// VerificationFile -
type VerificationFile struct {
	*postgres.Table[*storage.VerificationFile]
}

// NewVerificationFile -
func NewVerificationFile(db *database.Bun) *VerificationFile {
	return &VerificationFile{
		Table: postgres.NewTable[*storage.VerificationFile](db),
	}
}

// ByTaskId -
func (f *VerificationFile) ByTaskId(ctx context.Context, id uint64) (files []storage.VerificationFile, err error) {
	err = f.DB().NewSelect().
		Model((*storage.VerificationFile)(nil)).
		Where("verification_task_id = ?", id).
		Scan(ctx, &files)

	return
}

// BulkSave -
func (f *VerificationFile) BulkSave(ctx context.Context, files ...*storage.VerificationFile) error {
	if len(files) == 0 {
		return nil
	}
	_, err := f.DB().NewInsert().Model(&files).Exec(ctx)

	return err
}

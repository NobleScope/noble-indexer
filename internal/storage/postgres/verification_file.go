package postgres

import (
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

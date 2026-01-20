package storage

import (
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IVerificationFile interface {
	storage.Table[*VerificationFile]
}

// VerificationFile -
type VerificationFile struct {
	bun.BaseModel `bun:"verification_file" comment:"Table with contract verification task files."`

	Id                 uint64 `bun:",pk,notnull,autoincrement" comment:"Unique internal identity"`
	File               []byte `bun:"file,type:bytea"           comment:"File"`
	VerificationTaskId uint64 `bun:"verification_task_id"      comment:"Verification task id"`
}

// TableName -
func (VerificationFile) TableName() string {
	return "verification_file"
}

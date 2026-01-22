package storage

import (
	"context"
	"time"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IVerificationTask interface {
	storage.Table[*VerificationTask]

	Latest(ctx context.Context) (VerificationTask, error)
	ByContractId(ctx context.Context, contractId uint64) (VerificationTask, error)
}

// VerificationTask -
type VerificationTask struct {
	bun.BaseModel `bun:"verification_task" comment:"Table with contract verification tasks"`

	Id                  uint64                       `bun:",pk,autoincrement"                               comment:"Unique internal identity"`
	Status              types.VerificationTaskStatus `bun:"status,type:verification_task_status"            comment:"Verification task status"`
	CreationTime        time.Time                    `bun:"creation_time,default:now()"                     comment:"Task creation time"`
	CompletionTime      time.Time                    `bun:"completion_time"                                 comment:"Task completion time"`
	ContractId          uint64                       `bun:"contract_id,unique:verification_contract_id_idx" comment:"Contract id"`
	CompilerVersion     string                       `bun:"compiler_version,notnull"         comment:"Compiler version"`
	LicenseType         types.LicenseType            `bun:"license_type,type:license_type" comment:"License type"`
	OptimizationEnabled *bool                        `bun:"optimization_enabled"                comment:"Optimization enabled"`
	OptimizationRuns    *uint                        `bun:"optimization_runs"                comment:"Optimization runs"`
}

// TableName -
func (VerificationTask) TableName() string {
	return "verification_task"
}

package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upTraceCallType, downTraceCallType)
}

func upTraceCallType(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `ALTER TABLE public."trace" ADD COLUMN IF NOT EXISTS "call_type" text`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `COMMENT ON COLUMN public."trace"."call_type" IS 'Call type (call, delegatecall, staticcall, callcode)'`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `UPDATE public."trace" SET call_type = type WHERE type IN ('delegatecall', 'staticcall')`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `UPDATE public."trace" SET type = 'call' WHERE type IN ('delegatecall', 'staticcall')`); err != nil {
		return err
	}
	return nil
}

func downTraceCallType(ctx context.Context, db *bun.DB) error {
	return nil
}

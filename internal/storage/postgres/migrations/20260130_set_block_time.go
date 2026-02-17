package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upSetBlockTime, downSetBlockTime)
}

func upSetBlockTime(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
	with _data as (
		select 
		height, 
		(extract('epoch' from time) - extract('epoch' from LAG(time) OVER (ORDER BY height))) * 1000 as block_time
		from block 
		where height > 0 
		order by time desc
	)
	update block_stats 
	set block_time = _data.block_time 
	from _data
	where block_stats.height = _data.height
	`)
	return err
}

func downSetBlockTime(ctx context.Context, db *bun.DB) error {
	return nil
}

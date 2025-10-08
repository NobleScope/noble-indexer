package node

import (
	"context"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type Api interface {
	Head(ctx context.Context) (pkgTypes.Level, error)
	Block(ctx context.Context, level pkgTypes.Level) (pkgTypes.Block, error)
	BlockBulk(ctx context.Context, levels ...pkgTypes.Level) ([]pkgTypes.BlockData, error)
}

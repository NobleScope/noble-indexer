package storage

import (
	"context"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/pkg/errors"
)

func saveERC4337UserOps(
	ctx context.Context,
	tx storage.Transaction,
	userOps []*storage.UserOp,
	txHashes map[string]uint64,
	addresses map[string]uint64,
) error {
	if len(userOps) == 0 {
		return nil
	}

	for i := range userOps {
		txId, ok := txHashes[userOps[i].Tx.Hash.String()]
		if !ok {
			return errors.Errorf("can't find tx hash key: %s", userOps[i].Tx.Hash.String())
		}
		userOps[i].TxId = txId

		senderId, ok := addresses[userOps[i].Sender.Hash.String()]
		if !ok {
			return errors.Errorf("can't find sender addr key: %s", userOps[i].Sender.Hash.String())
		}
		userOps[i].SenderId = senderId

		bundlerId, ok := addresses[userOps[i].Bundler.String()]
		if !ok {
			return errors.Errorf("can't find bundler addr key: %s", userOps[i].Bundler.String())
		}
		userOps[i].BundlerId = bundlerId

		if userOps[i].Paymaster != nil {
			paymasterId, paymasterExist := addresses[userOps[i].Paymaster.String()]
			if !paymasterExist {
				return errors.Errorf("can't find paymaster addr key: %s", userOps[i].Paymaster.String())
			}
			userOps[i].PaymasterId = &paymasterId
		}
	}

	return tx.SaveERC4337UserOps(ctx, userOps...)
}

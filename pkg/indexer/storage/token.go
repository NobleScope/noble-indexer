package storage

import (
	"context"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/pkg/errors"
)

func saveTokens(
	ctx context.Context,
	tx storage.Transaction,
	tokens []*storage.Token,
	addresses map[string]uint64,
) (int64, error) {
	if len(tokens) == 0 {
		return 0, nil
	}

	for i := range tokens {
		id, ok := addresses[tokens[i].Contract.Address.String()]
		if !ok {
			return 0, errors.Errorf("can't find contract key: %s", tokens[i].Contract.Address.String())
		}
		tokens[i].ContractId = id
	}

	totalTokens, err := tx.SaveTokens(ctx, tokens...)
	return totalTokens, err
}

package token_metadata

import (
	"context"
	"path"
	"time"

	"github.com/baking-bad/noble-indexer/internal/ipfs"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/postgres"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/baking-bad/noble-indexer/pkg/node"
	"github.com/baking-bad/noble-indexer/pkg/node/rpc"
	tmTypes "github.com/baking-bad/noble-indexer/pkg/token_metadata/types"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/pkg/errors"
)

type Module struct {
	modules.BaseModule

	api         node.Api
	pool        ipfs.Pool
	pg          postgres.Storage
	storage     sdk.Transactable
	syncPeriod  time.Duration
	retryDelay  time.Duration
	cfg         config.Config
	abiRegistry *ABIRegistry
}

func NewModule(pg postgres.Storage, cfg config.Config) *Module {
	pool, err := ipfs.New(cfg.TokenMetadataResolver.MetadataGateways)
	if err != nil {
		panic(err)
	}

	var nodeRpc rpc.API
	if ds, ok := cfg.DataSources["node_rpc"]; ok && ds.URL != "" {
		nodeRpc = rpc.NewApi(ds, rpc.WithTimeout(time.Second*time.Duration(ds.Timeout)), rpc.WithRateLimit(ds.RequestsPerSecond))
	}

	module := &Module{
		BaseModule: modules.New("token_metadata_resolver"),
		api:        &nodeRpc,
		pg:         pg,
		storage:    pg.Transactable,
		pool:       pool,
		cfg:        cfg,
		syncPeriod: time.Second * time.Duration(cfg.TokenMetadataResolver.SyncPeriod),
		retryDelay: time.Minute * time.Duration(cfg.TokenMetadataResolver.RetryDelay),
	}

	module.abiRegistry = &ABIRegistry{
		abi: make(map[string]abi.ABI),
	}

	if err = module.abiRegistry.LoadInterfaces(path.Join(cfg.Indexer.AssetsDir, "abi")); err != nil {
		panic(err)
	}

	return module
}

func (m *Module) Close() error {
	m.Log.Info().Msg("closing module...")
	m.G.Wait()

	return nil
}

func (m *Module) Start(ctx context.Context) {
	m.Log.Info().Msg("starting module...")
	m.G.GoCtx(ctx, m.receive)
}

func (m *Module) receive(ctx context.Context) {
	if err := m.sync(ctx); err != nil {
		m.Log.Err(err).Msg("sync")
	}

	ticker := time.NewTicker(m.syncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.sync(ctx); err != nil {
				m.Log.Err(err).Msg("sync")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Module) sync(ctx context.Context) error {
	ts, err := m.pg.Token.PendingMetadata(ctx, m.retryDelay, m.cfg.TokenMetadataResolver.RequestBulkSize)
	if err != nil {
		return errors.Wrap(err, "get tokens")
	}

	m.Log.Info().Int("count", len(ts)).Msg("new tokens received")
	tokens := make(map[uint64]*storage.Token)
	tokenMetadata := make([]pkgTypes.TokenMetadataRequest, 0)
	for _, t := range ts {
		contract, err := m.pg.Contracts.GetByID(ctx, t.ContractId)
		if err != nil {
			m.Log.Err(err).Uint64("contract ID", t.ContractId).Msg("failed to get contract by id")
			continue
		}
		iABI, ok := m.abiRegistry.abi[t.Type.String()]
		if !ok {
			m.Log.Err(err).Str("token type", t.Type.String()).Msg("no abi for token type")
			continue
		}

		tokens[t.Id] = t
		tokenMetadata = append(tokenMetadata, pkgTypes.TokenMetadataRequest{
			Id:        t.Id,
			Address:   contract.Address.Address,
			ABI:       iABI,
			Interface: t.Type,
			TokenID:   t.TokenID.BigInt(),
		})
	}

	if len(tokenMetadata) == 0 {
		return nil
	}

	metadata, err := m.api.TokenMetadataBulk(ctx, tokenMetadata)
	if err != nil {
		return errors.Wrap(err, "token metadata bulk")
	}

	err = m.resolveMetadata(ctx, tokens, metadata)
	if err != nil {
		return errors.Wrap(err, "resolve metadata")
	}

	updatedTokens := make([]*storage.Token, 0)
	for _, t := range tokens {
		updatedTokens = append(updatedTokens, t)
	}

	if err := m.save(ctx, updatedTokens); err != nil {
		m.Log.Err(err).Msg("save")
	}

	return err
}

func (m *Module) resolveMetadata(ctx context.Context, tokens map[uint64]*storage.Token, metadata map[uint64]pkgTypes.TokenMetadata) error {
	for i, t := range metadata {
		iABI, ok := m.abiRegistry.abi[tokens[i].Type.String()]
		if !ok {
			err := errors.Errorf("no abi for token type: %s", tokens[i].Type.String())
			m.Log.Err(err).
				Str("contract", tokens[i].Contract.String()).
				Msg("no abi for this token type")

			m.failMetadata(tokens[i], err)
			continue
		}

		switch tokens[i].Type {
		case types.ERC20:
			var name string
			err := iABI.UnpackIntoInterface(&name, tmTypes.Name.String(), t.Name)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].Name = name

			var symbol string
			err = iABI.UnpackIntoInterface(&symbol, tmTypes.Symbol.String(), t.Symbol)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].Symbol = symbol

			var decimals uint8
			err = iABI.UnpackIntoInterface(&decimals, tmTypes.Decimals.String(), t.Decimals)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].Decimals = decimals
		case types.ERC721:
			var name string
			err := iABI.UnpackIntoInterface(&name, tmTypes.Name.String(), t.Name)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].Name = name

			var symbol string
			err = iABI.UnpackIntoInterface(&symbol, tmTypes.Symbol.String(), t.Symbol)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].Symbol = symbol

			var metadataLink string
			err = iABI.UnpackIntoInterface(&metadataLink, tmTypes.TokenUri.String(), t.URI)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].MetadataLink = metadataLink

			md, err := m.pool.LoadMetadata(ctx, metadataLink)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}

			tokens[i].Metadata = md

		case types.ERC1155:
			var metadataLink string
			err := iABI.UnpackIntoInterface(&metadataLink, tmTypes.Uri.String(), t.URI)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].MetadataLink = metadataLink

			md, err := m.pool.LoadMetadata(ctx, metadataLink)
			if err != nil {
				m.failMetadata(tokens[i], err)
				continue
			}
			tokens[i].Metadata = md

		default:
			continue
		}

		tokens[i].Status = types.Success
		tokens[i].Error = ""
	}

	return nil
}

func (m *Module) failMetadata(token *storage.Token, err error) {
	m.Log.Err(err).
		Str("token_ID", token.TokenID.String()).
		Str("token_type", token.Type.String()).
		Str("metadata_link", token.MetadataLink).
		Msg(err.Error())

	token.RetryCount += 1
	if token.RetryCount >= m.cfg.TokenMetadataResolver.RetryCount {
		token.Status = types.Failed
		m.Log.Err(err).Msg("retry limit exceeded. Status set to failed")
	}
	token.Error = err.Error()
}

func (m *Module) save(ctx context.Context, tokens []*storage.Token) error {
	tx, err := postgres.BeginTransaction(ctx, m.storage)
	if err != nil {
		return err
	}
	defer tx.Close(ctx)

	err = tx.SaveTokenMetadata(ctx, tokens...)
	if err != nil {
		return errors.Wrap(err, "save token metadata")
	}
	m.Log.Info().Int("count", len(tokens)).Msg("token metadata saved")

	if err := tx.Flush(ctx); err != nil {
		return tx.HandleError(ctx, err)
	}

	return nil
}

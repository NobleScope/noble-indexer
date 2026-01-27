package parser

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/baking-bad/noble-indexer/pkg/indexer/config"
	"github.com/pkg/errors"

	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Module struct {
	modules.BaseModule

	cfg config.Indexer
	abi map[types.TokenType]abi.ABI
}

var _ modules.Module = (*Module)(nil)

const (
	InputName  = "blocks"
	OutputName = "data"
	StopOutput = "stop"
)

func NewModule(cfg config.Indexer) Module {
	m := Module{
		BaseModule: modules.New("parser"),
		cfg:        cfg,
		abi:        make(map[types.TokenType]abi.ABI),
	}

	err := m.createABIs()
	if err != nil {
		panic(err)
	}

	m.CreateInputWithCapacity(InputName, 128)
	m.CreateOutput(OutputName)
	m.CreateOutput(StopOutput)

	return m
}

func (p *Module) Start(ctx context.Context) {
	p.Log.Info().Msg("starting parser module...")
	p.G.GoCtx(ctx, p.listen)
}

func (p *Module) Close() error {
	p.Log.Info().Msg("closing...")
	p.G.Wait()
	return nil
}

func (p *Module) createABIs() error {
	erc20ABIJson, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), string(types.ERC20))
	if err != nil {
		return errors.Wrap(err, "reading erc20 abi json")
	}
	erc721ABIJson, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), string(types.ERC721))
	if err != nil {
		return errors.Wrap(err, "reading erc721 abi json")
	}
	erc1155ABIJson, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), string(types.ERC1155))
	if err != nil {
		return errors.Wrap(err, "reading erc1155 abi json")
	}
	eip4337EntrypointV06Json, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), "EIP4337_entrypoint_v06.json")
	if err != nil {
		return errors.Wrap(err, "reading EIP4337_entrypoint_v06 abi json")
	}
	eip4337EntrypointV07Json, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), "EIP4337_entrypoint_v07.json")
	if err != nil {
		return errors.Wrap(err, "reading EIP4337_entrypoint_v07 abi json")
	}
	eip4337EntrypointForPaymasterJson, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), "EIP4337_entrypoint_for_paymaster.json")
	if err != nil {
		return errors.Wrap(err, "reading ERC4337_entrypoint_for_paymaster abi json")
	}
	eip4337AccountV06Json, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), "EIP4337_account_v06.json")
	if err != nil {
		return errors.Wrap(err, "reading EIP4337_account_v06 abi json")
	}
	eip4337AccountV07Json, err := readJson(filepath.Join(p.cfg.AssetsDir, "abi"), "EIP4337_account_v07.json")
	if err != nil {
		return errors.Wrap(err, "reading EIP4337_account_v07 abi json")
	}

	erc20ABI, err := abi.JSON(bytes.NewReader(erc20ABIJson))
	if err != nil {
		return errors.Wrap(err, "parsing erc20 abi")
	}
	erc721ABI, err := abi.JSON(bytes.NewReader(erc721ABIJson))
	if err != nil {
		return errors.Wrap(err, "parsing erc721 abi")
	}
	erc1155ABI, err := abi.JSON(bytes.NewReader(erc1155ABIJson))
	if err != nil {
		return errors.Wrap(err, "parsing erc1155 abi")
	}
	eip4337EntrypointV06ABI, err := abi.JSON(bytes.NewReader(eip4337EntrypointV06Json))
	if err != nil {
		return errors.Wrap(err, "parsing EIP4337_entrypoint_v06 abi")
	}
	eip4337EntrypointV07ABI, err := abi.JSON(bytes.NewReader(eip4337EntrypointV07Json))
	if err != nil {
		return errors.Wrap(err, "parsing EIP4337_entrypoint_v07 abi")
	}
	eip4337EntrypointForPaymasterABI, err := abi.JSON(bytes.NewReader(eip4337EntrypointForPaymasterJson))
	if err != nil {
		return errors.Wrap(err, "parsing EIP4337_entrypoint_for_paymaster abi")
	}
	eip4337AccountV06ABI, err := abi.JSON(bytes.NewReader(eip4337AccountV06Json))
	if err != nil {
		return errors.Wrap(err, "parsing EIP4337_account_v06 abi")
	}
	eip4337AccountV07ABI, err := abi.JSON(bytes.NewReader(eip4337AccountV07Json))
	if err != nil {
		return errors.Wrap(err, "parsing EIP4337_account_v07 abi")
	}

	p.abi[types.ERC20] = erc20ABI
	p.abi[types.ERC721] = erc721ABI
	p.abi[types.ERC1155] = erc1155ABI
	p.abi["EIP4337_entrypoint_v06"] = eip4337EntrypointV06ABI
	p.abi["EIP4337_entrypoint_v07"] = eip4337EntrypointV07ABI
	p.abi["EIP4337_entrypoint_for_paymaster"] = eip4337EntrypointForPaymasterABI
	p.abi["EIP4337_account_v06"] = eip4337AccountV06ABI
	p.abi["EIP4337_account_v07"] = eip4337AccountV07ABI

	return nil
}

func readJson(path, filename string) ([]byte, error) {
	if filepath.Ext(filename) != ".json" {
		filename += ".json"
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(wd, path, filename)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	return data, nil
}

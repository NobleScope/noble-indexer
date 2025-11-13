package receiver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/baking-bad/noble-indexer/pkg/types"
)

func (r *Module) receiveGenesis(ctx context.Context) error {
	r.Log.Info().Msg("receiving genesis")
	file, err := readGenesisFile(r.cfg.AssetsDir)
	if err != nil {
		return err
	}

	var genesis types.Genesis
	if err = json.Unmarshal(file, &genesis); err != nil {
		return err
	}
	if err != nil {
		return err
	}

	r.MustOutput(GenesisOutput).Push(genesis)
	genesisDoneInput := r.MustInput(GenesisDoneInput)

	// Wait until the genesis block will be saved
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-genesisDoneInput.Listen():
			return nil
		}
	}
}

func readGenesisFile(path string) ([]byte, error) {
	wd, _ := os.Getwd()
	p := filepath.Join(wd, path, "genesis.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	return data, nil
}

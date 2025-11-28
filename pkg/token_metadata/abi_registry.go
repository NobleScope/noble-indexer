package token_metadata

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type ABIRegistry struct {
	abi map[string]abi.ABI
}

func (r *ABIRegistry) LoadInterfaces(dir string) error {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file %s: %w", path, err)
		}
		defer file.Close()

		iABI, err := abi.JSON(file)
		if err != nil {
			return fmt.Errorf("parse ABI %s: %w", path, err)
		}

		name := strings.TrimSuffix(filepath.Base(path), ".json")
		r.abi[name] = iABI

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

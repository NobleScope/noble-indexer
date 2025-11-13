package metadata_resolver

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type ABIInterface struct {
	Name string
	ABI  *abi.ABI
}

type ABIRegistry struct {
	interfaces []ABIInterface
}

func (r *ABIRegistry) LoadInterfaces(dir string) error {
	var interfaces []ABIInterface

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
		interfaces = append(interfaces, ABIInterface{
			Name: name,
			ABI:  &iABI,
		})

		return nil
	})

	if err != nil {
		return err
	}

	r.interfaces = interfaces
	return nil
}

func (r *ABIRegistry) MatchABI(contractABIData json.RawMessage) ([]string, error) {
	var implInterfaces []string

	contractABI, err := abi.JSON(strings.NewReader(string(contractABIData)))
	if err != nil {
		return implInterfaces, fmt.Errorf("invalid contract ABI: %w", err)
	}

	for _, i := range r.interfaces {
		if implementsInterface(&contractABI, i.ABI) {
			implInterfaces = append(implInterfaces, i.Name)
		}
	}

	return implInterfaces, nil
}

func implementsInterface(contractABI, interfaceABI *abi.ABI) bool {
	for name, iMethod := range interfaceABI.Methods {
		contractMethod, ok := contractABI.Methods[name]
		if !ok {
			return false
		}

		if !matchSignatures(contractMethod, iMethod) {
			return false
		}
	}
	return true
}

func matchSignatures(a, b abi.Method) bool {
	if a.Name != b.Name {
		return false
	}

	if len(a.Inputs) != len(b.Inputs) {
		return false
	}

	for i := range a.Inputs {
		if a.Inputs[i].Type.String() != b.Inputs[i].Type.String() {
			return false
		}
	}

	if len(a.Outputs) != len(b.Outputs) {
		return false
	}

	for i := range a.Outputs {
		if a.Outputs[i].Type.String() != b.Outputs[i].Type.String() {
			return false
		}
	}

	return true
}

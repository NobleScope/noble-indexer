package parser

import (
	"bytes"
	"fmt"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/pkg/errors"
	"github.com/unpackdev/solgo/bytecode"
)

const VersionUnknown = "unknown"

func ParseEvmContractMetadata(contract *storage.Contract) error {
	if !hasContractMetadata(contract.Code) {
		version, err := extractCompilerVersion(contract.Code)
		if err != nil {
			return err
		}
		contract.CompilerVersion = version
		return nil
	}

	metadata, err := bytecode.DecodeContractMetadata(contract.Code)
	if err != nil {
		return err
	}

	contract.CompilerVersion = metadata.GetCompilerVersion()
	contract.MetadataLink = metadata.GetIPFS()

	return nil
}

func extractCBORSection(b []byte) ([]byte, error) {
	if len(b) < 2 {
		return nil, errors.Errorf("too short bytecode")
	}

	cborLen := int(b[len(b)-2])<<8 | int(b[len(b)-1])
	if cborLen <= 0 || cborLen > len(b)-2 {
		return nil, errors.Errorf("invalid cbor length: %d", cborLen)
	}

	cborStart := len(b) - 2 - cborLen
	if cborStart < 0 {
		return nil, errors.Errorf("invalid cbor start")
	}

	return b[cborStart : len(b)-2], nil
}

func hasContractMetadata(b []byte) bool {
	cborData, err := extractCBORSection(b)
	if err != nil {
		return false
	}

	sigSwarm := []byte{0xa1, 0x65, 0x62, 0x7a, 0x7a, 0x72}
	sigIPFS := []byte{0xa2, 0x64, 0x69, 0x70, 0x66, 0x73, 0x58}

	return bytes.Contains(cborData, sigSwarm) || bytes.Contains(cborData, sigIPFS)
}

func extractCompilerVersion(b []byte) (string, error) {
	cborData, err := extractCBORSection(b)
	if err != nil {
		return VersionUnknown, nil
	}

	pos := bytes.Index(cborData, []byte("solc"))
	if pos == -1 {
		return VersionUnknown, nil
	}

	if pos+5 >= len(cborData) {
		return VersionUnknown, nil
	}

	versionBytes := cborData[pos+5:]
	if len(versionBytes) < 3 {
		return VersionUnknown, nil
	}

	major := int(versionBytes[0])
	minor := int(versionBytes[1])
	patch := int(versionBytes[2])

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

package contract_verifier

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/lmittmann/go-solc"
	"github.com/pkg/errors"
)

const (
	defaultOptimizationRuns = 200
)

var contractNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type BytecodeParts struct {
	main     []byte
	metadata []byte
}

type VerificationResult struct {
	ABI             json.RawMessage
	CompilerVersion string
	Language        string
}

func (m *Module) verify(ctx context.Context, task storage.VerificationTask, files []storage.VerificationFile) (*VerificationResult, error) {
	if len(files) == 0 {
		m.Log.Err(errors.New("verification files not found")).Uint64("contract_id", task.ContractId).Msg("verification contract does not contain any files")
		return nil, errors.New("no source code files found for verification task")
	}

	if !contractNameRe.MatchString(task.ContractName) {
		return nil, errors.Errorf("invalid contract name: %s", task.ContractName)
	}

	contract, err := m.pg.Contracts.GetByID(ctx, task.ContractId)
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("could not get contract")
		return nil, errors.Wrap(err, "get contract by id")
	}

	var evmVersion string
	if task.EVMVersion != nil {
		evmVersion = task.EVMVersion.String()
		m.Log.Info().Str("evm_version", evmVersion).Msg("using EVM version from task")
	} else {
		evmVersion = detectEVMVersion(contract.Code.Bytes()).String()
		m.Log.Info().Str("evm_version", evmVersion).Msg("auto-detected EVM version from onchain bytecode")
	}

	compilerVersion := strings.TrimPrefix(task.CompilerVersion, "v")
	if idx := strings.Index(compilerVersion, "+"); idx != -1 {
		compilerVersion = compilerVersion[:idx]
	}

	compiler := solc.New(solc.Version(compilerVersion))

	var opts []solc.Option
	opts = append(opts, solc.WithEVMVersion(solc.EVMVersion(evmVersion)))

	if task.OptimizationEnabled != nil && *task.OptimizationEnabled {
		runs := uint64(defaultOptimizationRuns)
		if task.OptimizationRuns != nil {
			runs = uint64(*task.OptimizationRuns)
		}
		opts = append(opts, solc.WithOptimizer(&solc.Optimizer{
			Enabled: true,
			Runs:    runs,
		}))
	}

	if task.ViaIR {
		opts = append(opts, solc.WithViaIR(true))
	}

	tmpDir1, err := os.MkdirTemp("", "contract-verify-1-")
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("could not create temporary directory")
		return nil, errors.Wrap(err, "create temp directory for first compilation")
	}
	defer func() {
		_ = os.RemoveAll(tmpDir1)
	}()

	mainContractFileName := task.ContractName + ".sol"
	sources := buildSourceMap(files)

	if err := writeSourceFiles(tmpDir1, sources); err != nil {
		return nil, err
	}

	contract1, err := compiler.Compile(tmpDir1, task.ContractName, opts...)
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("failed to compile contract")
		return nil, errors.Wrap(err, "compile contract")
	}

	if len(contract1.Runtime) == 0 {
		m.Log.Error().Uint64("contract_id", task.ContractId).Msg("contract does not contain any runtimes")
		return nil, errors.New("compilation produced no runtime bytecode")
	}

	m.Log.Info().
		Uint64("task_id", task.Id).
		Int("files_count", len(files)).
		Msg("contract compiled successfully")

	tmpDir2, err := os.MkdirTemp("", "contract-verify-2-")
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("could not create temporary directory")
		return nil, errors.Wrap(err, "create temp directory for second compilation")
	}
	defer func() {
		_ = os.RemoveAll(tmpDir2)
	}()

	modifiedSources := make(map[string]string, len(sources))
	for path, content := range sources {
		if filepath.Base(path) == mainContractFileName {
			content += "\n"
		}
		modifiedSources[path] = content
	}

	if err := writeSourceFiles(tmpDir2, modifiedSources); err != nil {
		return nil, err
	}

	contract2, err := compiler.Compile(tmpDir2, task.ContractName, opts...)
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("failed to compile modified contract")
		return nil, errors.New("failed to compile modified contract for metadata detection")
	}

	runtimeParts := splitBytecode(contract1.Runtime, contract2.Runtime)
	contractBytes := contract.Code.Bytes()
	if len(contractBytes) < len(runtimeParts.main) {
		m.Log.Error().Uint64("contract_id", task.ContractId).Msg("onchain bytecode is shorter than compiled main part")
		return nil, errors.Errorf("onchain bytecode is shorter than compiled main part: %d < %d",
			len(contractBytes), len(runtimeParts.main))
	}

	contractMainPart := contractBytes[:len(runtimeParts.main)]
	if !bytecodeEqualIgnoringImmutables(runtimeParts.main, contractMainPart) {
		m.Log.Error().
			Str("compiled_main", hex.EncodeToString(runtimeParts.main)).
			Str("onchain_main", hex.EncodeToString(contractMainPart)).
			Msg("main bytecode parts do not match")
		return nil, errors.New("bytecode verification failed: main parts do not match")
	}

	m.Log.Info().
		Uint64("contract_id", task.ContractId).
		Msg("bytecode verification successfully: main parts match")

	abiJSON, err := json.Marshal(contract1.ABI)
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("failed to marshal ABI")
		return nil, errors.Wrap(err, "marshal contract ABI")
	}

	parsedABI, err := abi.JSON(bytes.NewReader(abiJSON))
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("failed to parse ABI")
		return nil, errors.Wrap(err, "parse contract ABI")
	}

	if contract.TxId == nil {
		m.Log.Error().Uint64("contract_id", task.ContractId).Msg("deployment transaction ID should not be nil")
		return nil, errors.New("deployment transaction not found")
	}

	deployTx, err := m.pg.Tx.GetByID(ctx, *contract.TxId)
	if err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("failed to get deployment transaction")
		return nil, errors.Wrap(err, "get deployment transaction")
	}

	if err := m.verifyConstructorArgs(parsedABI, deployTx.Input, len(contract1.Constructor)); err != nil {
		m.Log.Err(err).Uint64("contract_id", task.ContractId).Msg("failed to verify contract constructor args")
		return nil, errors.Wrap(err, "verify constructor arguments")
	}

	return &VerificationResult{
		ABI:             abiJSON,
		CompilerVersion: task.CompilerVersion,
		Language:        "Solidity",
	}, nil
}

// detectEVMVersion detects the minimum required EVM version based on opcodes present in bytecode.
// It scans for opcodes introduced in different EVM upgrades and returns the newest required version.
func detectEVMVersion(bytecode []byte) types.EVMVersion {
	detectedVersion := types.Istanbul // default safe version

	// Opcode -> minimum EVM version that introduced it
	// We check from newest to oldest, so we can break early
	for i := 0; i < len(bytecode); {
		b := bytecode[i]

		// Skip PUSH operands: PUSH1 (0x60) to PUSH32 (0x7f) are followed by 1-32 bytes of data
		if b >= 0x60 && b <= 0x7f {
			i += int(b-0x60) + 2
			continue
		}
		i++

		switch b {
		// Cancun opcodes (EIP-1153, EIP-5656, EIP-4844)
		case 0x5c, 0x5d: // TLOAD, TSTORE
			return types.Cancun
		case 0x5e: // MCOPY
			return types.Cancun
		case 0x49, 0x4a: // BLOBHASH, BLOBBASEFEE
			return types.Cancun

		// Shanghai opcodes (EIP-3855)
		case 0x5f: // PUSH0
			if detectedVersion < types.Shanghai {
				detectedVersion = types.Shanghai
			}

		// London opcodes (EIP-3198)
		case 0x48: // BASEFEE
			if detectedVersion < types.London {
				detectedVersion = types.London
			}

		// Istanbul opcodes (EIP-1344, EIP-1884)
		case 0x46, 0x47: // CHAINID, SELFBALANCE
			if detectedVersion < types.Istanbul {
				detectedVersion = types.Istanbul
			}

		// Constantinople opcodes (EIP-145, EIP-1014, EIP-1052)
		case 0x1b, 0x1c, 0x1d: // SHL, SHR, SAR
			if detectedVersion < types.Constantinople {
				detectedVersion = types.Constantinople
			}
		case 0x3f: // EXTCODEHASH
			if detectedVersion < types.Constantinople {
				detectedVersion = types.Constantinople
			}
		case 0xf5: // CREATE2
			if detectedVersion < types.Constantinople {
				detectedVersion = types.Constantinople
			}

		// Byzantium opcodes (EIP-140, EIP-211, EIP-214)
		case 0xfd: // REVERT
			if detectedVersion < types.Byzantium {
				detectedVersion = types.Byzantium
			}
		case 0x3d, 0x3e: // RETURNDATASIZE, RETURNDATACOPY
			if detectedVersion < types.Byzantium {
				detectedVersion = types.Byzantium
			}
		case 0xfa: // STATICCALL
			if detectedVersion < types.Byzantium {
				detectedVersion = types.Byzantium
			}
		}
	}

	return detectedVersion
}

func (m *Module) verifyConstructorArgs(
	parsedABI abi.ABI,
	deployInput []byte,
	compiledDeployLen int,
) error {
	deployInputLen := len(deployInput)

	// If deploy input is shorter than compiled deploy code, the contract was likely deployed via a factory (CREATE/CREATE2)
	if deployInputLen < compiledDeployLen {
		m.Log.Info().
			Int("deploy_input_len", deployInputLen).
			Int("compiled_deploy_len", compiledDeployLen).
			Msg("contract deployed via factory, skipping constructor args verification")
		return nil
	}

	hasConstructorParams := len(parsedABI.Constructor.Inputs) > 0

	if hasConstructorParams {
		constructorArgs := deployInput[compiledDeployLen:]
		decodedArgs, err := parsedABI.Constructor.Inputs.Unpack(constructorArgs)
		if err != nil {
			return errors.Wrap(err, "decode constructor arguments")
		}
		m.Log.Info().
			Int("arg_count", len(decodedArgs)).
			Msg("constructor arguments decoded successfully")
	} else {
		extraBytes := deployInputLen - compiledDeployLen
		if extraBytes >= 32 {
			return errors.Errorf("deployment input has %d extra bytes after deploy code, but constructor has no parameters", extraBytes)
		}
		m.Log.Info().Msg("no constructor parameters verified")
	}

	return nil
}

// bytecodeEqualIgnoringImmutables compares compiled and onchain runtime bytecodes,
// treating PUSH32 (0x7f) followed by 32 zero bytes in compiled as immutable variable
// placeholders that can have any value in onchain bytecode.
func bytecodeEqualIgnoringImmutables(compiled, onchain []byte) bool {
	if len(compiled) != len(onchain) {
		return false
	}
	for i := 0; i < len(compiled); {
		if compiled[i] != onchain[i] {
			return false
		}
		// PUSH32 followed by 32 zero bytes = immutable placeholder, skip value
		if compiled[i] == 0x7f && i+32 < len(compiled) {
			isImmutable := true
			for j := 1; j <= 32; j++ {
				if compiled[i+j] != 0x00 {
					isImmutable = false
					break
				}
			}
			if isImmutable {
				i += 33
				continue
			}
		}
		i++
	}
	return true
}

func splitBytecode(bytecode1, bytecode2 []byte) BytecodeParts {
	minLen := len(bytecode1)
	if len(bytecode2) < minLen {
		minLen = len(bytecode2)
	}

	metadataStart := minLen
	for i := 0; i < minLen; i++ {
		if bytecode1[i] != bytecode2[i] {
			metadataStart = i
			break
		}
	}

	// CBOR metadata starts after 0xfe (INVALID opcode)
	// Find the last 0xfe before the divergence point and exclude it from main
	for i := metadataStart - 1; i >= 0; i-- {
		if bytecode1[i] == 0xfe {
			metadataStart = i + 1
			break
		}
	}

	return BytecodeParts{
		main:     bytecode1[:metadataStart],
		metadata: bytecode1[metadataStart:],
	}
}

var importRe = regexp.MustCompile(`import\s[^"';]*["']([^"']+)["']`)

// buildSourceMap builds a solc source map (key -> content) from uploaded files.
// Keys are determined by analyzing import paths in the source code:
//   - Files referenced by absolute imports get the full import path as key.
//   - Relative imports are resolved iteratively from known file positions.
//   - Files not referenced by any import get their basename as key.
func buildSourceMap(files []storage.VerificationFile) map[string]string {
	contentByName := make(map[string]string)
	for _, f := range files {
		contentByName[filepath.Base(f.Name)] = string(f.File)
	}

	pathSets := make(map[string]map[string]struct{})
	addPath := func(basename, path string) bool {
		if pathSets[basename] == nil {
			pathSets[basename] = make(map[string]struct{})
		}
		if _, exists := pathSets[basename][path]; exists {
			return false
		}
		pathSets[basename][path] = struct{}{}
		return true
	}

	// Step 1: collect absolute (non-relative) imports
	for _, f := range files {
		for _, match := range importRe.FindAllSubmatch(f.File, -1) {
			importPath := string(match[1])
			if strings.HasPrefix(importPath, ".") {
				continue
			}
			addPath(filepath.Base(importPath), importPath)
		}
	}

	// Step 2: iteratively resolve relative imports from known positions
	changed := true
	for changed {
		changed = false
		for _, f := range files {
			fileName := filepath.Base(f.Name)
			paths, ok := pathSets[fileName]
			if !ok {
				continue
			}
			for fp := range paths {
				fileDir := filepath.Dir(fp)
				for _, match := range importRe.FindAllSubmatch(f.File, -1) {
					importPath := string(match[1])
					if !strings.HasPrefix(importPath, ".") {
						continue
					}
					resolved := filepath.Clean(filepath.Join(fileDir, importPath))
					if addPath(filepath.Base(resolved), resolved) {
						changed = true
					}
				}
			}
		}
	}

	sources := make(map[string]string)
	for basename, content := range contentByName {
		paths := pathSets[basename]
		if len(paths) == 0 {
			sources[basename] = content
		} else {
			for p := range paths {
				sources[p] = content
			}
		}
	}
	return sources
}

func writeSourceFiles(dir string, sources map[string]string) error {
	cleanDir := filepath.Clean(dir) + string(os.PathSeparator)
	for path, content := range sources {
		filePath := filepath.Join(dir, path)
		if !strings.HasPrefix(filePath, cleanDir) {
			return errors.Errorf("path traversal detected: %s", path)
		}
		if subDir := filepath.Dir(filePath); subDir != dir {
			if err := os.MkdirAll(subDir, 0755); err != nil {
				return errors.Wrapf(err, "create directory for %s", path)
			}
		}
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			return errors.Wrapf(err, "write source file %s", path)
		}
	}
	return nil
}

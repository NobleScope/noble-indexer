package contract_verifier

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/lmittmann/go-solc"
	"github.com/pkg/errors"
)

const (
	defaultOptimizationRuns = 200
)

type BytecodeParts struct {
	main     []byte
	metadata []byte
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

	return BytecodeParts{
		main:     bytecode1[:metadataStart],
		metadata: bytecode1[metadataStart:],
	}
}

func (m *Module) verify(ctx context.Context, task storage.VerificationTask, files []storage.VerificationFile) error {
	if len(files) == 0 {
		return errors.New("no source code files found for verification task")
	}

	sourceCode := files[0].File // todo: temporarily
	tmpDir, err := os.MkdirTemp("", "contract-verify-")
	if err != nil {
		return errors.Wrap(err, "create temp directory")
	}

	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	contractFile := filepath.Join(tmpDir, "contract.sol")
	if err := os.WriteFile(contractFile, sourceCode, 0644); err != nil {
		return errors.Wrap(err, "write source code file")
	}

	compilerVersion := strings.TrimPrefix(task.CompilerVersion, "v")
	if idx := strings.Index(compilerVersion, "+"); idx != -1 {
		compilerVersion = compilerVersion[:idx]
	}

	compiler := solc.New(solc.Version(compilerVersion))

	var opts []solc.Option
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

	contract, err := compiler.Compile(tmpDir, task.ContractName, opts...)
	if err != nil {
		m.Log.Err(err).Msg("failed to compile contract")
		return errors.Wrap(err, "compile contract")
	}

	if len(contract.Runtime) == 0 {
		return errors.New("compilation produced no runtime bytecode")
	}

	m.Log.Info().
		Uint64("task_id", task.Id).
		Int("runtime_bytecode_size", len(contract.Runtime)).
		Int("constructor_bytecode_size", len(contract.Constructor)).
		Msg("contract compiled successfully")

	modifiedSourceCode := append(sourceCode, []byte("\n")...)
	contractFile2 := filepath.Join(tmpDir, "contract2.sol")
	if err := os.WriteFile(contractFile2, modifiedSourceCode, 0644); err != nil {
		return errors.Wrap(err, "write modified source code file")
	}

	contract2, err := compiler.Compile(tmpDir, task.ContractName, opts...)
	if err != nil {
		m.Log.Warn().Err(err).Msg("failed to compile modified contract for metadata detection, skipping split")
		contract2 = nil
	}

	onchainContract, err := m.pg.Contracts.GetByID(ctx, task.ContractId)
	if err != nil {
		return errors.Wrap(err, "get contract by id")
	}

	var parts BytecodeParts
	if contract2 != nil && len(contract2.Runtime) > 0 {
		parts = splitBytecode(contract.Runtime, contract2.Runtime)
		m.Log.Info().
			Int("main_size", len(parts.main)).
			Int("metadata_size", len(parts.metadata)).
			Msg("bytecode split into main and metadata parts")
	} else {
		parts = BytecodeParts{
			main:     contract.Runtime,
			metadata: []byte{},
		}
	}

	onchainBytes := onchainContract.Code.Bytes()
	if len(onchainBytes) < len(parts.main) {
		return errors.Errorf("on-chain bytecode is shorter than compiled main part: %d < %d",
			len(onchainBytes), len(parts.main))
	}

	onchainMain := onchainBytes[:len(parts.main)]
	if !bytes.Equal(parts.main, onchainMain) {
		m.Log.Error().
			Str("compiled_main", hex.EncodeToString(parts.main)).
			Str("onchain_main", hex.EncodeToString(onchainMain)).
			Msg("main bytecode parts do not match")
		return errors.New("bytecode verification failed: main parts do not match")
	}

	m.Log.Info().
		Uint64("contract_id", task.ContractId).
		Int("main_part_size", len(parts.main)).
		Msg("bytecode verification successfully: main parts match")

	abiJSON, err := json.Marshal(contract.ABI)
	if err != nil {
		return errors.Wrap(err, "marshal contract ABI")
	}

	parsedABI, err := abi.JSON(bytes.NewReader(abiJSON))
	if err != nil {
		return errors.Wrap(err, "parse contract ABI")
	}

	if len(parsedABI.Constructor.Inputs) > 0 {
		if onchainContract.TxId == nil {
			return errors.New("contract has constructor parameters but no deployment transaction found")
		}

		deployTx, err := m.pg.Tx.GetByID(ctx, *onchainContract.TxId)
		if err != nil {
			return errors.Wrap(err, "get deployment transaction")
		}

		if len(deployTx.Input) < len(contract.Constructor) {
			return errors.New("deployment transaction input is shorter than constructor bytecode")
		}

		constructorArgs := deployTx.Input[len(contract.Constructor):]
		if len(constructorArgs) == 0 {
			return errors.New("constructor has parameters but no arguments found in deployment transaction")
		}

		decodedArgs, err := parsedABI.Constructor.Inputs.Unpack(constructorArgs)
		if err != nil {
			return errors.Wrap(err, "decode constructor arguments")
		}

		m.Log.Info().
			Int("arg_count", len(decodedArgs)).
			Msg("constructor arguments decoded successfully")
	} else {
		m.Log.Info().Msg("no constructor parameters, skipping constructor args verification")
	}

	// todo: save verified contract, update verification task status and remove verification_files
	// todo: update indexer state (inc qty of verified contracts)

	return nil
}

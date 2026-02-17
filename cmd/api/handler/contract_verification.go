package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/postgres"
	storageTypes "github.com/NobleScope/noble-indexer/internal/storage/types"
	"github.com/NobleScope/noble-indexer/pkg/types"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const MaxFileSize = 10 * 1024 * 1024 // 10 MB

var (
	contractNameRe    = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	compilerVersionRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+(\+commit\.[0-9a-f]+)?$`)
)

type uploadedSourceFile struct {
	name    string
	content []byte
}

type ContractVerificationHandler struct {
	contract storage.IContract
	task     storage.IVerificationTask
	file     storage.IVerificationFile
	beginTx  func(context.Context) (storage.Transaction, error)
}

func NewContractVerificationHandler(
	contract storage.IContract,
	task storage.IVerificationTask,
	file storage.IVerificationFile,
	transactable sdk.Transactable,
) *ContractVerificationHandler {
	return &ContractVerificationHandler{
		contract: contract,
		task:     task,
		file:     file,
		beginTx: func(ctx context.Context) (storage.Transaction, error) {
			return postgres.BeginTransaction(ctx, transactable)
		},
	}
}

type verificationResponse struct {
	Result string `json:"result"`
}

// ContractVerify godoc
//
//	@Summary		Creates a task to verify the specified contract
//	@Description	Creates a task to verify the specified contract with source code files. Multiple .sol files can be uploaded.
//	@Tags			verification
//	@ID				contract-verification
//	@Param			contract_address	formData string true  "Contract address"
//	@Param			contract_name       formData string true  "Contract name in Solidity source"
//	@Param			source_code         formData file   true  "Source code files (.sol). Multiple files allowed."
//	@Param			compiler_version    formData string true  "Compiler version"
//	@Param			license_type		formData string true  "License type"										Enums(none, unlicense, mit, gnu_gpl_v2, gnu_gpl_v3, gnu_lgpl_v2_1, gnu_lgpl_v3, bsd_2_clause, bsd_3_clause, mpl_2_0, osl_3_0, apache_2_0, gnu_agpl_v3, bsl_1_1)
//	@Param			optimization_enabled formData bool   false "Optimization enabled"
//	@Param			optimization_runs   formData int    false "Optimization runs"
//	@Param			evm_version         formData string false "EVM version. Auto-detected if not specified."		Enums(homestead, tangerineWhistle, spuriousDragon, byzantium, constantinople, petersburg, istanbul, berlin, london, paris, shanghai, cancun, prague)
//	@Param			via_ir              formData bool   false "Compile via Yul IR pipeline"
//	@Accept			multipart/form-data
//	@Produce		json
//	@Success		200	{object}	verificationResponse
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/verification/code [post]
func (handler *ContractVerificationHandler) ContractVerify(c echo.Context) error {
	address := c.FormValue("contract_address")
	if address == "" {
		return badRequestError(c, errors.New("contract address is required"))
	}

	contractName := c.FormValue("contract_name")
	if contractName == "" {
		return badRequestError(c, errors.New("contract name is required"))
	}
	if !contractNameRe.MatchString(contractName) {
		return badRequestError(c, errors.New("invalid contract name: must match [a-zA-Z_][a-zA-Z0-9_]*"))
	}

	compilerVersion := c.FormValue("compiler_version")
	if compilerVersion == "" {
		return badRequestError(c, errors.New("compiler version is required"))
	}
	if !compilerVersionRe.MatchString(compilerVersion) {
		return badRequestError(c, errors.New("invalid compiler version: must be semver (e.g. 0.8.20)"))
	}

	licenseTypeStr := c.FormValue("license_type")
	if licenseTypeStr == "" {
		return badRequestError(c, errors.New("license type is required"))
	}

	licenseType, err := storageTypes.ParseLicenseType(licenseTypeStr)
	if err != nil {
		return badRequestError(c, errors.Wrap(err, "invalid license type"))
	}

	var optimizationEnabled *bool
	var optimizationRuns *uint

	if optEnabledStr := c.FormValue("optimization_enabled"); optEnabledStr != "" {
		optEnabled, err := strconv.ParseBool(optEnabledStr)
		if err != nil {
			return badRequestError(c, errors.Wrap(err, "invalid optimization_enabled value"))
		}
		optimizationEnabled = &optEnabled
	}

	if optRunsStr := c.FormValue("optimization_runs"); optRunsStr != "" {
		optRuns64, err := strconv.ParseUint(optRunsStr, 10, 32)
		if err != nil {
			return badRequestError(c, errors.Wrap(err, "invalid optimization_runs value"))
		}
		optRuns := uint(optRuns64)
		optimizationRuns = &optRuns
	}

	var evmVersion *storageTypes.EVMVersion
	if evmVersionStr := c.FormValue("evm_version"); evmVersionStr != "" {
		v, err := storageTypes.ParseEVMVersion(evmVersionStr)
		if err != nil {
			return badRequestError(c, errors.Wrap(err, "invalid EVM version"))
		}
		evmVersion = &v
	}

	var viaIR bool
	if viaIRStr := c.FormValue("via_ir"); viaIRStr != "" {
		viaIR, err = strconv.ParseBool(viaIRStr)
		if err != nil {
			return badRequestError(c, errors.Wrap(err, "invalid via_ir value"))
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		return badRequestError(c, errors.New("failed to parse multipart form"))
	}

	fileHeaders := form.File["source_code"]
	if len(fileHeaders) == 0 {
		return badRequestError(c, errors.New("at least one source code file is required"))
	}

	sourceFiles := make([]uploadedSourceFile, 0, len(fileHeaders))
	foundPragma := false

	for _, fileHeader := range fileHeaders {
		if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".sol") {
			return badRequestError(c, errors.Errorf("only .sol files are allowed: %s", fileHeader.Filename))
		}

		file, err := fileHeader.Open()
		if err != nil {
			return badRequestError(c, errors.Wrapf(err, "failed to open file %s", fileHeader.Filename))
		}

		content, err := io.ReadAll(io.LimitReader(file, MaxFileSize+1))
		_ = file.Close()
		if err != nil {
			return badRequestError(c, errors.Wrapf(err, "failed to read file %s", fileHeader.Filename))
		}

		if len(content) > MaxFileSize {
			return badRequestError(c, errors.Errorf("file %s is too large, maximum size is 10 MB", fileHeader.Filename))
		}

		if len(content) == 0 {
			return badRequestError(c, errors.Errorf("file %s is empty", fileHeader.Filename))
		}

		if bytes.Contains(content, []byte("pragma solidity")) {
			foundPragma = true
		}

		sourceFiles = append(sourceFiles, uploadedSourceFile{
			name:    fileHeader.Filename,
			content: content,
		})
	}

	if !foundPragma {
		return badRequestError(c, errors.New("at least one file must contain 'pragma solidity'"))
	}

	hash, err := types.HexFromString(address)
	if err != nil {
		return badRequestError(c, err)
	}

	contract, err := handler.contract.ByHash(c.Request().Context(), hash)
	if err != nil {
		if handler.contract.IsNoRows(err) {
			return badRequestError(c, errors.New("contract not found"))
		}
		return handleError(c, err, handler.contract)
	}

	tasks, err := handler.task.ByContractId(c.Request().Context(), contract.Id)
	if err != nil {
		if !handler.task.IsNoRows(err) {
			return handleError(c, err, handler.task)
		}
	}

	for i := range tasks {
		if tasks[i].ContractId != 0 &&
			(tasks[i].Status == storageTypes.VerificationStatusNew || tasks[i].Status == storageTypes.VerificationStatusPending) {
			return badRequestError(c, errors.New("such a contract is already in the verification process"))
		}
		if tasks[i].ContractId != 0 && tasks[i].Status == storageTypes.VerificationStatusSuccess {
			return badRequestError(c, errors.New("such a contract is already verified"))
		}
	}

	newTask := storage.VerificationTask{
		Status:              storageTypes.VerificationStatusNew,
		ContractId:          contract.Id,
		ContractName:        contractName,
		CompilerVersion:     compilerVersion,
		LicenseType:         licenseType,
		OptimizationEnabled: optimizationEnabled,
		OptimizationRuns:    optimizationRuns,
		EVMVersion:          evmVersion,
		ViaIR:               viaIR,
	}

	ctx := c.Request().Context()

	tx, err := handler.beginTx(ctx)
	if err != nil {
		return handleError(c, err, handler.task)
	}
	defer tx.Close(ctx)

	if err := tx.AddVerificationTask(ctx, &newTask); err != nil {
		return handleError(c, err, handler.task)
	}

	files := make([]*storage.VerificationFile, 0, len(sourceFiles))
	for i := range sourceFiles {
		files = append(files, &storage.VerificationFile{
			Name:               sourceFiles[i].name,
			File:               sourceFiles[i].content,
			VerificationTaskId: newTask.Id,
		})
	}

	if err := tx.SaveVerificationFiles(ctx, files...); err != nil {
		return handleError(c, err, handler.file)
	}

	if err := tx.Flush(ctx); err != nil {
		return handleError(c, err, handler.task)
	}

	return c.JSON(http.StatusOK, verificationResponse{Result: "success"})
}

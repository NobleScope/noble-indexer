package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/baking-bad/noble-indexer/internal/storage"
	ts "github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type ContractVerificationHandler struct {
	contract storage.IContract
	task     storage.IVerificationTask
	file     storage.IVerificationFile
}

func NewContractVerificationHandler(
	contract storage.IContract,
	task storage.IVerificationTask,
	file storage.IVerificationFile,
) *ContractVerificationHandler {
	return &ContractVerificationHandler{
		contract: contract,
		task:     task,
		file:     file,
	}
}

type verificationResponse struct {
	Result string `json:"result"`
}

// ContractVerify godoc
//
//	@Summary		Creates a task to verify the specified contract
//	@Description	Creates a task to verify the specified contract with source code file
//	@Tags			verification
//	@ID				contract-verification
//	@Param			contract            formData string true  "Contract address"
//	@Param			source_code         formData file   true  "Source code file"
//	@Param			compiler_version    formData string true  "Compiler version"
//	@Param			license_type		formData string true  "License type"
//	@Param			optimization_enabled formData bool  false "Optimization enabled"
//	@Param			optimization_runs   formData int    false "Optimization runs"
//	@Accept			multipart/form-data
//	@Produce		json
//	@Success		200	{object}	verificationResponse
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/verification/code [post]
func (handler *ContractVerificationHandler) ContractVerify(c echo.Context) error {
	address := c.FormValue("contract")
	if address == "" {
		return badRequestError(c, errors.New("contract address is required"))
	}

	compilerVersion := c.FormValue("compiler_version")
	if compilerVersion == "" {
		return badRequestError(c, errors.New("compiler version is required"))
	}

	licenseType := c.FormValue("license_type")
	if licenseType == "" {
		return badRequestError(c, errors.New("license type is required"))
	}

	fileHeader, err := c.FormFile("source_code")
	if err != nil {
		return badRequestError(c, errors.New("source code file is required"))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return badRequestError(c, errors.Wrap(err, "failed to open source code file"))
	}
	defer func() {
		_ = file.Close()
	}()

	sourceCode, err := io.ReadAll(file)
	if err != nil {
		return badRequestError(c, errors.Wrap(err, "failed to read source code file"))
	}

	if len(sourceCode) == 0 {
		return badRequestError(c, errors.New("source code file is empty"))
	}

	if !bytes.Contains(sourceCode, []byte("pragma solidity")) {
		return badRequestError(c, errors.New("the uploaded file is not the source code of the contract"))
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

	task, err := handler.task.ByContractId(c.Request().Context(), contract.Id)
	if err != nil {
		if !handler.task.IsNoRows(err) {
			return handleError(c, err, handler.task)
		}
	}
	if task.ContractId != 0 {
		return badRequestError(c, errors.New("such a contract is already in the verification process"))
	}

	newTask := storage.VerificationTask{
		Status:     ts.VerificationStatusNew,
		ContractId: contract.Id,
	}
	err = handler.task.Save(c.Request().Context(), &newTask)
	if err != nil {
		return handleError(c, err, handler.task)
	}

	verificationFile := storage.VerificationFile{
		File:               sourceCode,
		VerificationTaskId: newTask.Id,
	}
	err = handler.file.Save(c.Request().Context(), &verificationFile)
	if err != nil {
		return handleError(c, err, handler.file)
	}

	return c.JSON(http.StatusOK, verificationResponse{Result: "success"})
}

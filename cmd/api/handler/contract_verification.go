package handler

import (
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

type postVerificationTaskRequest struct {
	Contract            string `json:"contract"             validate:"required,address"`
	SourceCode          string `json:"source_code"          validate:"required"`
	CompilerVersion     string `json:"compiler_version"     validate:"required"`
	LicenseType         string `json:"license_type"         validate:"required"`
	OptimizationEnabled bool   `json:"optimization_enabled"`
	OptimizationRuns    uint   `json:"optimization_runs"`
}

type verificationResponse struct {
	Result string `json:"result"`
}

// ContractVerify godoc
//
//	@Summary		Creates a task to verify the specified contract
//	@Description	Creates a task to verify the specified contract
//	@Tags			verification
//	@ID				contract-verification
//	@Param			request	body postVerificationTaskRequest true "Request body containing contract, source_code, compiler_version and license_type"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	verificationResponse
//	@Failure		400	{object}	Error
//	@Failure		500	{object}	Error
//	@Router			/verification/code [post]
func (handler *ContractVerificationHandler) ContractVerify(c echo.Context) error {
	req, err := bindAndValidate[postVerificationTaskRequest](c)
	if err != nil {
		return badRequestError(c, err)
	}
	hash, err := types.HexFromString(req.Contract)
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

	return c.JSON(http.StatusOK, verificationResponse{Result: "success"})
}

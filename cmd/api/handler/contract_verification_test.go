package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/mock"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ContractVerificationTestSuite struct {
	suite.Suite

	echo     *echo.Echo
	ctrl     *gomock.Controller
	contract *mock.MockIContract
	task     *mock.MockIVerificationTask
	file     *mock.MockIVerificationFile
	handler  *ContractVerificationHandler
}

func (s *ContractVerificationTestSuite) SetupTest() {
	s.echo = echo.New()
	s.echo.Validator = NewApiValidator()

	s.ctrl = gomock.NewController(s.T())
	s.contract = mock.NewMockIContract(s.ctrl)
	s.task = mock.NewMockIVerificationTask(s.ctrl)
	s.file = mock.NewMockIVerificationFile(s.ctrl)

	s.handler = NewContractVerificationHandler(s.contract, s.task, s.file)
}

func (s *ContractVerificationTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestSuiteContractVerification_Run(t *testing.T) {
	suite.Run(t, new(ContractVerificationTestSuite))
}

// createMultipartRequest builds a multipart/form-data request with the given fields and files.
func createMultipartRequest(fields map[string]string, files map[string][]byte) (*http.Request, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, err
		}
	}

	for name, content := range files {
		part, err := writer.CreateFormFile("source_code", name)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(content); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func (s *ContractVerificationTestSuite) validFields() map[string]string {
	return map[string]string{
		"contract_address": testAddressHex3.String(),
		"contract_name":    "TestContract",
		"compiler_version": "0.8.20",
		"license_type":     "mit",
	}
}

func (s *ContractVerificationTestSuite) validFiles() map[string][]byte {
	return map[string][]byte{
		"TestContract.sol": []byte("// SPDX-License-Identifier: MIT\npragma solidity ^0.8.20;\ncontract TestContract {}"),
	}
}

// --- Required field validation ---

func (s *ContractVerificationTestSuite) TestContractVerify_MissingAddress() {
	fields := s.validFields()
	delete(fields, "contract_address")

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_MissingContractName() {
	fields := s.validFields()
	delete(fields, "contract_name")

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_MissingCompilerVersion() {
	fields := s.validFields()
	delete(fields, "compiler_version")

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_MissingLicenseType() {
	fields := s.validFields()
	delete(fields, "license_type")

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_InvalidLicenseType() {
	fields := s.validFields()
	fields["license_type"] = "invalid_license"

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// --- Optional field validation ---

func (s *ContractVerificationTestSuite) TestContractVerify_InvalidOptimizationEnabled() {
	fields := s.validFields()
	fields["optimization_enabled"] = "not_a_bool"

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_InvalidOptimizationRuns() {
	fields := s.validFields()
	fields["optimization_runs"] = "not_a_number"

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_InvalidEVMVersion() {
	fields := s.validFields()
	fields["evm_version"] = "invalid_version"

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_InvalidViaIR() {
	fields := s.validFields()
	fields["via_ir"] = "not_a_bool"

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// --- File validation ---

func (s *ContractVerificationTestSuite) TestContractVerify_NoFiles() {
	fields := s.validFields()

	req, err := createMultipartRequest(fields, nil)
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_NonSolFile() {
	fields := s.validFields()
	files := map[string][]byte{
		"contract.txt": []byte("pragma solidity ^0.8.20;"),
	}

	req, err := createMultipartRequest(fields, files)
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_EmptyFile() {
	fields := s.validFields()
	files := map[string][]byte{
		"TestContract.sol": {},
	}

	req, err := createMultipartRequest(fields, files)
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_NoPragma() {
	fields := s.validFields()
	files := map[string][]byte{
		"TestContract.sol": []byte("contract TestContract {}"),
	}

	req, err := createMultipartRequest(fields, files)
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// --- Contract lookup ---

func (s *ContractVerificationTestSuite) TestContractVerify_InvalidAddress() {
	fields := s.validFields()
	fields["contract_address"] = "not_a_hex"

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_ContractNotFound() {
	fields := s.validFields()

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(storage.Contract{}, sql.ErrNoRows)
	s.contract.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)
}

// --- Existing task checks ---

func (s *ContractVerificationTestSuite) TestContractVerify_AlreadyInProgress() {
	fields := s.validFields()

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(testContract, nil)

	s.task.EXPECT().
		ByContractId(gomock.Any(), testContract.Id).
		Return([]storage.VerificationTask{
			{ContractId: testContract.Id, Status: types.VerificationStatusNew},
		}, nil)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var resp Error
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Require().Contains(resp.Message, "already in the verification process")
}

func (s *ContractVerificationTestSuite) TestContractVerify_AlreadyVerified() {
	fields := s.validFields()

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(testContract, nil)

	s.task.EXPECT().
		ByContractId(gomock.Any(), testContract.Id).
		Return([]storage.VerificationTask{
			{ContractId: testContract.Id, Status: types.VerificationStatusSuccess},
		}, nil)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var resp Error
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Require().Contains(resp.Message, "already verified")
}

// --- Success ---

func (s *ContractVerificationTestSuite) TestContractVerify_Success() {
	fields := s.validFields()

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(testContract, nil)

	s.task.EXPECT().
		ByContractId(gomock.Any(), testContract.Id).
		Return(nil, sql.ErrNoRows)
	s.task.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true)

	s.task.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, task *storage.VerificationTask) error {
			s.Require().Equal("TestContract", task.ContractName)
			s.Require().Equal("0.8.20", task.CompilerVersion)
			s.Require().Equal(types.VerificationStatusNew, task.Status)
			s.Require().Equal(testContract.Id, task.ContractId)
			task.Id = 1
			return nil
		})

	s.file.EXPECT().
		BulkSave(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, files ...*storage.VerificationFile) error {
			s.Require().Len(files, 1)
			s.Require().Equal("TestContract.sol", files[0].Name)
			s.Require().Equal(uint64(1), files[0].VerificationTaskId)
			return nil
		})

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var resp verificationResponse
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Require().Equal("success", resp.Result)
}

func (s *ContractVerificationTestSuite) TestContractVerify_SuccessWithOptionalParams() {
	fields := s.validFields()
	fields["optimization_enabled"] = "true"
	fields["optimization_runs"] = "500"
	fields["evm_version"] = "shanghai"
	fields["via_ir"] = "true"

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(testContract, nil)

	s.task.EXPECT().
		ByContractId(gomock.Any(), testContract.Id).
		Return(nil, sql.ErrNoRows)
	s.task.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true)

	s.task.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, task *storage.VerificationTask) error {
			s.Require().NotNil(task.OptimizationEnabled)
			s.Require().True(*task.OptimizationEnabled)
			s.Require().NotNil(task.OptimizationRuns)
			s.Require().Equal(uint(500), *task.OptimizationRuns)
			s.Require().NotNil(task.EVMVersion)
			s.Require().Equal(types.Shanghai, *task.EVMVersion)
			s.Require().True(task.ViaIR)
			task.Id = 1
			return nil
		})

	s.file.EXPECT().
		BulkSave(gomock.Any(), gomock.Any()).
		Return(nil)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_MultipleFiles() {
	fields := s.validFields()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range fields {
		s.Require().NoError(writer.WriteField(key, val))
	}

	mainContent := []byte("// SPDX-License-Identifier: MIT\npragma solidity ^0.8.20;\nimport \"./Helper.sol\";\ncontract TestContract {}")
	helperContent := []byte("// SPDX-License-Identifier: MIT\npragma solidity ^0.8.20;\nlibrary Helper {}")

	for name, content := range map[string][]byte{"TestContract.sol": mainContent, "Helper.sol": helperContent} {
		part, err := writer.CreateFormFile("source_code", name)
		s.Require().NoError(err)
		_, err = part.Write(content)
		s.Require().NoError(err)
	}
	s.Require().NoError(writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(testContract, nil)

	s.task.EXPECT().
		ByContractId(gomock.Any(), testContract.Id).
		Return(nil, sql.ErrNoRows)
	s.task.EXPECT().
		IsNoRows(sql.ErrNoRows).
		Return(true)

	s.task.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, task *storage.VerificationTask) error {
			task.Id = 1
			return nil
		})

	s.file.EXPECT().
		BulkSave(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, files ...*storage.VerificationFile) error {
			s.Require().Len(files, 2)
			return nil
		})

	err := s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *ContractVerificationTestSuite) TestContractVerify_FailedTaskExistsThenNewAllowed() {
	fields := s.validFields()

	req, err := createMultipartRequest(fields, s.validFiles())
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	s.contract.EXPECT().
		ByHash(gomock.Any(), testAddressHex3).
		Return(testContract, nil)

	s.task.EXPECT().
		ByContractId(gomock.Any(), testContract.Id).
		Return([]storage.VerificationTask{
			{ContractId: testContract.Id, Status: types.VerificationStatusFailed},
		}, nil)

	s.task.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, task *storage.VerificationTask) error {
			task.Id = 2
			return nil
		})

	s.file.EXPECT().
		BulkSave(gomock.Any(), gomock.Any()).
		Return(nil)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, rec.Code)

	var resp verificationResponse
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Require().Equal("success", resp.Result)
}

// --- File size validation ---

func (s *ContractVerificationTestSuite) TestContractVerify_FileTooLarge() {
	fields := s.validFields()
	largeContent := make([]byte, MaxFileSize+1)
	copy(largeContent, []byte("pragma solidity"))
	files := map[string][]byte{
		"Large.sol": largeContent,
	}

	req, err := createMultipartRequest(fields, files)
	s.Require().NoError(err)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	err = s.handler.ContractVerify(c)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusBadRequest, rec.Code)

	var resp Error
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Require().Contains(resp.Message, fmt.Sprintf("too large"))
}

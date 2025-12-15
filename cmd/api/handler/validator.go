package handler

import (
	"net/http"
	"regexp"

	"github.com/baking-bad/noble-indexer/internal/storage/types"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var evmAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
var evmTransactionHashRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)

type ApiValidator struct {
	validator *validator.Validate
}

func NewApiValidator() *ApiValidator {
	v := validator.New()
	if err := v.RegisterValidation("address", addressValidator()); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("tx_hash", txHashValidator()); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("trace_type", traceTypeValidator()); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("token_type", tokenTypeValidator()); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("transfer_type", transferTypeValidator()); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("proxy_contract_type", proxyContractTypeValidator()); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("proxy_contract_status", proxyContractStatusValidator()); err != nil {
		panic(err)
	}
	return &ApiValidator{validator: v}
}

func (v *ApiValidator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func addressValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		address := fl.Field().String()
		return evmAddressRegex.MatchString(address)
	}
}

func txHashValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		txHash := fl.Field().String()
		return evmTransactionHashRegex.MatchString(txHash)
	}
}

func traceTypeValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := types.ParseTraceType(fl.Field().String())
		return err == nil
	}
}

func tokenTypeValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := types.ParseTokenType(fl.Field().String())
		return err == nil
	}
}
func transferTypeValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := types.ParseTransferType(fl.Field().String())
		return err == nil
	}
}

func proxyContractTypeValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := types.ParseProxyType(fl.Field().String())
		return err == nil
	}
}

func proxyContractStatusValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := types.ParseProxyStatus(fl.Field().String())
		return err == nil
	}
}

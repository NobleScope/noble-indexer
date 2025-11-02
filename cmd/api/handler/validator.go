package handler

import (
	"net/http"
	"regexp"

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
	if err := v.RegisterValidation("txHash", txHashValidator()); err != nil {
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

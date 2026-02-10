package handler

import (
	"net/http"

	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/labstack/echo/v4"
)

type ConstantHandler struct{}

func NewConstantHandler() *ConstantHandler {
	return &ConstantHandler{}
}

// Enums godoc
//
//	@Summary		Get enumeration values
//	@Description	Returns all possible enumeration values used in the API including transaction types, transaction statuses, trace types, token types, transfer types, proxy contract types, and proxy contract statuses. Use these values for filtering in other API endpoints.
//	@Tags			general
//	@ID				get-enums
//	@Produce		json
//	@Success		200	{object}	responses.Enums	"All enumeration values available in the API"
//	@Router			/enums [get]
func (handler *ConstantHandler) Enums(c echo.Context) error {
	return c.JSON(http.StatusOK, responses.NewEnums())
}

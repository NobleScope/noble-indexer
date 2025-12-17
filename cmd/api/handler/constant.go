package handler

import (
	"net/http"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/labstack/echo/v4"
)

type ConstantHandler struct{}

func NewConstantHandler() *ConstantHandler {
	return &ConstantHandler{}
}

// Enums godoc
//
//	@Summary		Get noble enumerators
//	@Description	Get noble enumerators
//	@Tags			general
//	@ID				get-enums
//	@Produce		json
//	@Success		200	{object}	responses.Enums
//	@Router			/enums [get]
func (handler *ConstantHandler) Enums(c echo.Context) error {
	return c.JSON(http.StatusOK, responses.NewEnums())
}

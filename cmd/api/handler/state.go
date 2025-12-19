package handler

import (
	"net/http"

	"github.com/baking-bad/noble-indexer/cmd/api/handler/responses"
	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/labstack/echo/v4"
)

type StateHandler struct {
	state       storage.IState
	indexerName string
}

func NewStateHandler(state storage.IState, indexerName string) *StateHandler {
	return &StateHandler{
		state:       state,
		indexerName: indexerName,
	}
}

// Head godoc
//
//	@Summary		Get indexer state
//	@Description	Returns the current state of the blockchain indexer including the latest indexed block height and timestamp. Useful for checking indexer synchronization status.
//	@Tags			general
//	@ID				head
//	@Produce		json
//	@Success		200	{object}	responses.State	"Current indexer state"
//	@Success		204					"State not available"
//	@Failure		400	{object}	Error		"Invalid request"
//	@Failure		500	{object}	Error		"Internal server error"
//	@Router			/head [get]
func (sh *StateHandler) Head(c echo.Context) error {
	state, err := sh.state.ByName(c.Request().Context(), sh.indexerName)
	if err != nil {
		return handleError(c, err, sh.state)
	}

	return c.JSON(http.StatusOK, responses.NewState(state))
}

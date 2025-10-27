package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	errInvalidAddress = errors.New("invalid address")
	errCancelRequest  = "pq: canceling statement due to user request"
)

type NoRows interface {
	IsNoRows(err error) bool
}

type Error struct {
	Message string `json:"message"`
}

func badRequestError(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, Error{
		Message: err.Error(),
	})
}

func internalServerError(c echo.Context, err error) error {
	return c.JSON(http.StatusInternalServerError, Error{
		Message: err.Error(),
	})
}

func handleError(c echo.Context, err error, noRows NoRows) error {
	if err == nil {
		return nil
	}
	if err.Error() == errCancelRequest {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return c.JSON(http.StatusRequestTimeout, Error{
			Message: "timeout",
		})
	}
	if errors.Is(err, context.Canceled) {
		return c.JSON(http.StatusBadGateway, Error{
			Message: err.Error(),
		})
	}
	if noRows.IsNoRows(err) {
		return c.NoContent(http.StatusNoContent)
	}
	if errors.Is(err, errInvalidAddress) {
		return badRequestError(c, err)
	}
	return internalServerError(c, err)
}

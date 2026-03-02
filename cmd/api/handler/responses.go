package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CursorResponse wraps list results with an optional cursor for pagination.
//
//	@Description	Paginated list response with cursor
type CursorResponse struct {
	Result any    `json:"result"`
	Cursor string `json:"cursor,omitempty"`
}

func returnArray[T any](c echo.Context, arr []T) error {
	if arr == nil {
		return c.JSON(http.StatusOK, []any{})
	}

	return c.JSON(http.StatusOK, arr)
}

func returnCursorList[T any](c echo.Context, arr []T, cursor string) error {
	var result any = arr
	if arr == nil {
		result = []any{}
	}
	return c.JSON(http.StatusOK, CursorResponse{
		Result: result,
		Cursor: cursor,
	})
}

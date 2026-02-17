package handler

import (
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/labstack/echo/v4"
)

type StatsHandler struct {
	state       storage.IState
	blockStats  storage.IBlockStats
	indexerName string
}

func NewStatsHandler(state storage.IState, blockStats storage.IBlockStats, indexerName string) *StatsHandler {
	return &StatsHandler{
		state:       state,
		blockStats:  blockStats,
		indexerName: indexerName,
	}
}

// AvgBlockTime godoc
//
//	@Summary		Get average block time
//	@Description	Returns the average block time over a specified period. Useful for analyzing blockchain performance and trends.
//	@Tags			stats
//	@ID				avg-block-time
//	@Produce		json
//	@Success		200		{float}	    float64					"Average block time response"
//	@Failure		500		{object}	Error					"Internal server error"
//	@Router			/stats/block_time [get]
func (sh *StatsHandler) AvgBlockTime(c echo.Context) error {
	state, err := sh.state.ByName(c.Request().Context(), sh.indexerName)
	if err != nil {
		return handleError(c, err, sh.state)
	}
	blockTime, err := sh.blockStats.AvgBlockTime(c.Request().Context(), state.LastTime.Add(-3*time.Hour).UTC())
	if err != nil {
		return handleError(c, err, sh.blockStats)
	}
	return c.JSON(200, blockTime)
}

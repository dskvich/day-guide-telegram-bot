package handler

import (
	"log/slog"
	"net/http"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

func ImportHolidays() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type importSuccessResponse struct {
			Status          string `json:"status"`
			Message         string `json:"message"`
			ImportedRecords int    `json:"imported_records,omitempty"`
		}
		type importErrorResponse struct {
			Status    string `json:"status"`
			Message   string `json:"message"`
			Error     string `json:"error"`
			ErrorCode int    `json:"error_code"`
		}

		resp := importSuccessResponse{
			Status:          "success",
			Message:         "Holiday data imported successfully.",
			ImportedRecords: 120,
		}
		if err := encode(w, r, http.StatusOK, resp); err != nil {
			slog.Error("sending response", logger.Err(err))
		}
	})
}

package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type HolidayRepository interface {
	BatchInsert(ctx context.Context, holidays []domain.Holiday) error
}

type FileParser interface {
	ParseHolidays() ([]domain.Holiday, error)
}

func ImportHolidays(parser HolidayParser, repo HolidayRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type importSuccessResponse struct {
			Status          string `json:"status"`
			Message         string `json:"message"`
			ImportedRecords int    `json:"imported_records,omitempty"`
		}
		type importErrorResponse struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			errResponse := importErrorResponse{
				Status:  "error",
				Message: "failed getting form file",
				Error:   err.Error(),
			}
			if err := encode(w, r, http.StatusBadRequest, errResponse); err != nil {
				slog.Error("", logger.Err(err))
			}
			return
		}
		defer file.Close()

		slog.Debug("File Received", "file", fileHeader.Filename)

		date, err := parser.ParseDateFromFilename(fileHeader.Filename)
		if err != nil {
			errResponse := importErrorResponse{
				Status:  "error",
				Message: "failed parsing date from filename",
				Error:   err.Error(),
			}
			if err := encode(w, r, http.StatusInternalServerError, errResponse); err != nil {
				slog.Error("", logger.Err(err))
			}
			return
		}

		slog.Debug("Date Parsed From Filename", "file", fileHeader.Filename, "date", date)

		holidayNames, err := parser.ParseHolidays(file)
		if err != nil {
			errResponse := importErrorResponse{
				Status:  "error",
				Message: "failed parsing the HTML file",
				Error:   err.Error(),
			}
			if err := encode(w, r, http.StatusInternalServerError, errResponse); err != nil {
				slog.Error("", logger.Err(err))
			}
			return
		}

		slog.Debug("Holidays Parsed", "file", fileHeader.Filename, "holidays_count", len(holidayNames))

		holidays := make([]domain.Holiday, 0, len(holidayNames))
		for i, name := range holidayNames {
			holidays = append(holidays, domain.Holiday{
				OrderNumber: i,
				Name:        name,
				Date:        date,
			})
		}

		if err := repo.BatchInsert(r.Context(), holidays); err != nil {
			errResponse := importErrorResponse{
				Status:  "error",
				Message: "failed saving holidays to DB",
				Error:   err.Error(),
			}
			if err := encode(w, r, http.StatusInternalServerError, errResponse); err != nil {
				slog.Error("", logger.Err(err))
			}
			return
		}

		resp := importSuccessResponse{
			Status:          "success",
			Message:         fmt.Sprintf("%s file imported successfully", fileHeader.Filename),
			ImportedRecords: len(holidays),
		}
		if err := encode(w, r, http.StatusOK, resp); err != nil {
			slog.Error("sending response", logger.Err(err))
		}
	})
}

type HolidayParser struct{}

func (p *HolidayParser) ParseHolidays(r io.Reader) (map[int]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	holidays := make(map[int]string)
	doc.Find("span[itemprop='text']").Each(func(i int, s *goquery.Selection) {
		// Assign the index and the text directly to the map
		holidays[i] = s.Text()
	})
	return holidays, nil
}

// ParseDateFromFilename extracts a date from a filename formatted as YYYY-MM-DD.html
func (p *HolidayParser) ParseDateFromFilename(filename string) (time.Time, error) {
	trimmed := strings.TrimSuffix(filename, ".html")
	parts := strings.Split(trimmed, "-")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid filename format")
	}

	date, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

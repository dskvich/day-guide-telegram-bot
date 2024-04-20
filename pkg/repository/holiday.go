package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type holidayRepository struct {
	db *sql.DB
}

func NewHolidayRepository(db *sql.DB) *holidayRepository {
	return &holidayRepository{db: db}
}

func (repo *holidayRepository) BatchInsert(ctx context.Context, holidays []domain.Holiday) error {
	if len(holidays) == 0 {
		return nil // No holidays to insert
	}

	columns := []string{"name", "date"}
	placeholders := make([]string, 0, len(holidays))
	var args []interface{}

	for _, h := range holidays {
		placeholder := make([]string, len(columns))
		for j := range columns {
			args = append(args, h.Name, h.Date)
			placeholder[j] = fmt.Sprintf("$%d", len(args)-len(columns)+j+1)
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(placeholder, ", ")))
	}

	q := `INSERT INTO holidays (` + strings.Join(columns, ", ") + `) VALUES ` + strings.Join(placeholders, ",")

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("executing batch insert query: %v", err)
	}

	return nil
}

func (repo *holidayRepository) FetchByDate(ctx context.Context, date time.Time) ([]domain.Holiday, error) {
	q := `
		select name
		  from holidays
		 where date = $1;
	`

	// Format date to YYYY-MM-DD for comparison in SQL
	dateFormatted := date.Format("2006-01-02")

	rows, err := repo.db.QueryContext(ctx, q, dateFormatted)
	if err != nil {
		return nil, fmt.Errorf("querying holidays: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Warn("Failed to close rows", logger.Err(err))
		}
	}()

	var holidays []domain.Holiday
	for rows.Next() {
		var holiday domain.Holiday
		if err := rows.Scan(
			&holiday.Name,
		); err != nil {
			return nil, fmt.Errorf("scanning rows: %v", err)
		}

		holiday.Date = date
		holidays = append(holidays, holiday)
	}

	return holidays, rows.Err()
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/lib/pq"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type holidayRepository struct {
	db *sql.DB
}

func NewHolidayRepository(db *sql.DB) *holidayRepository {
	return &holidayRepository{db: db}
}

func (repo *holidayRepository) FetchByDate(ctx context.Context, date time.Time) ([]domain.Holiday, error) {
	q := `
        select h.order_number,
               h.name,
               ARRAY_AGG(c.name) AS categories
		from holidays h
		join holiday_category_links hcl ON h.id = hcl.holiday_id
		join holiday_categories c ON hcl.category_id = c.id
		where h.date = $1
		group by h.id
		order by h.order_number;
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
		var categoryNames []string

		if err := rows.Scan(
			&holiday.OrderNumber,
			&holiday.Name,
			pq.Array(&categoryNames),
		); err != nil {
			return nil, fmt.Errorf("scanning rows: %v", err)
		}

		holiday.Date = date

		for _, categoryName := range categoryNames {
			holiday.Categories = append(holiday.Categories, categoryName)
		}

		holidays = append(holidays, holiday)
	}

	return holidays, rows.Err()
}

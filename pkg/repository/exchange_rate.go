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

type exchangeRateRepository struct {
	db *sql.DB
}

func NewExchangeRateRepository(db *sql.DB) *exchangeRateRepository {
	return &exchangeRateRepository{db: db}
}

func (repo *exchangeRateRepository) Save(ctx context.Context, e *domain.ExchangeRate) error {
	set := []string{"base", "quote", "rate"}
	args := []any{e.Pair.Base, e.Pair.Quote, e.Rate}
	placeholder := []string{"?", "?", "?"}

	q := `insert into exchange_rates (` + strings.Join(set, ", ") + `) values (` + strings.Join(placeholder, ",") + `)`

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("executing query: %v", err)
	}

	return nil
}

func (repo *exchangeRateRepository) FetchLatestRate(ctx context.Context, pair domain.CurrencyPair) (*domain.ExchangeRate, error) {
	q := `
		select rate
		from exchange_rates
		where base = ?
		and quote = ?
		order by created_at desc
		limit 1;
	`

	e := domain.ExchangeRate{Pair: pair}
	if err := repo.db.QueryRowContext(ctx, q, pair.Base, pair.Quote).Scan(
		&e.Rate,
	); err != nil {
		return nil, fmt.Errorf("scanning row: %v", err)
	}

	return &e, nil
}

func (repo *exchangeRateRepository) FetchHistoryRate(ctx context.Context, pair domain.CurrencyPair, days int) ([]domain.ExchangeRate, error) {
	q := `
		WITH LastRateToday AS (
			SELECT strftime('%Y-%m-%d', created_at) AS date, 
			       rate
			FROM exchange_rates
			WHERE base = ? 
			  AND quote = ?
			  AND strftime('%Y-%m-%d', created_at) = strftime('%Y-%m-%d', 'now')
			ORDER BY created_at DESC
			LIMIT 1
		),
		 RecentAvgRates AS (
			 SELECT strftime('%Y-%m-%d', created_at) AS date, 
			        avg(rate) AS rate
			 FROM exchange_rates
			 WHERE base = ?
			   AND quote = ?
			   AND strftime('%Y-%m-%d', created_at) != strftime('%Y-%m-%d', 'now')
			 GROUP BY strftime('%Y-%m-%d', created_at)
			 ORDER BY date DESC
			 LIMIT ?
		 )
		SELECT * FROM RecentAvgRates
		UNION ALL
		SELECT * FROM LastRateToday
		ORDER BY date DESC
	`

	rows, err := repo.db.QueryContext(ctx, q, pair.Base, pair.Quote, pair.Base, pair.Quote, days)
	if err != nil {
		return nil, fmt.Errorf("querying history rate: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Warn("Failed to close rows", logger.Err(err))
		}
	}()

	var rates []domain.ExchangeRate
	for rows.Next() {
		var rate domain.ExchangeRate
		var tsStr string
		if err := rows.Scan(
			&tsStr,
			&rate.Rate,
		); err != nil {
			return nil, fmt.Errorf("scanning rows: %v", err)
		}

		layout := "2006-01-02" // replace with the actual layout of your timestamp
		timestamp, err := time.Parse(layout, tsStr)
		if err != nil {
			return nil, fmt.Errorf("parsing timestamp: %v", err)
		}
		rate.Timestamp = timestamp
		rates = append(rates, rate)
	}

	return rates, rows.Err()
}

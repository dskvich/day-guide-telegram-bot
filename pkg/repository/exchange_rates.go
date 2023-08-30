package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type exchangeRatesRepository struct {
	db *sql.DB
}

func NewExchangeRatesRepository(db *sql.DB) *exchangeRatesRepository {
	return &exchangeRatesRepository{db: db}
}

func (repo *exchangeRatesRepository) Save(ctx context.Context, e *domain.ExchangeRate) error {
	set := []string{"base", "quote", "rate"}
	args := []any{e.Pair.Base, e.Pair.Quote, e.Rate}
	placeholder := []string{"?", "?", "?"}

	q := `insert into exchange_rates (` + strings.Join(set, ", ") + `) values (` + strings.Join(placeholder, ",") + `)`

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("executing query: %v", err)
	}

	return nil
}

func (repo *exchangeRatesRepository) FetchLatestRate(ctx context.Context, pair domain.CurrencyPair) (*domain.ExchangeRate, error) {
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

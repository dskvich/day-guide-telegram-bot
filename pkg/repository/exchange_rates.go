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

func (repo *exchangeRatesRepository) Save(ctx context.Context, r *domain.USDExchangeRates) error {
	set := []string{"rub", "try"}
	args := []any{r.RUB, r.TRY}
	placeholder := []string{"?", "?"}

	q := `insert into usd_exchange_rates (` + strings.Join(set, ", ") + `) values (` + strings.Join(placeholder, ",") + `)`

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("executing query: %v", err)
	}

	return nil
}

func (repo *exchangeRatesRepository) FetchLatest(ctx context.Context) (*domain.USDExchangeRates, error) {
	q := `
		select 
		    rub,
		    try
		from usd_exchange_rates
		order by created_at desc
		limit 1;
	`

	var r domain.USDExchangeRates
	if err := repo.db.QueryRowContext(ctx, q).Scan(
		&r.RUB,
		&r.TRY,
	); err != nil {
		return nil, fmt.Errorf("scanning row: %v", err)
	}

	return &r, nil
}

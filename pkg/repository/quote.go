package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type quoteRepository struct {
	db *sql.DB
}

func NewQuoteRepository(db *sql.DB) *quoteRepository {
	return &quoteRepository{db: db}
}

func (repo *quoteRepository) Save(ctx context.Context, quote *domain.Quote) error {
	columns := []string{"quote", "author"}
	args := []any{quote.Quote, quote.Author}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	q := fmt.Sprintf("INSERT INTO quotes (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("executing query: %v", err)
	}

	return nil
}

func (repo *quoteRepository) FetchLatestQuote(ctx context.Context) (*domain.Quote, error) {
	q := `
		select 
		    quote,
			author
		from quotes
		order by created_at desc
		limit 1;
	`

	var quote domain.Quote
	if err := repo.db.QueryRowContext(ctx, q).Scan(
		&quote.Quote,
		&quote.Author,
	); err != nil {
		return nil, fmt.Errorf("scanning row: %v", err)
	}

	return &quote, nil
}

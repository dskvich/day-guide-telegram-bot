package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRateFetcher interface {
	FetchLatestRate(context.Context, domain.CurrencyPair) (*domain.ExchangeRate, error)
}

type ExchangeRateFormatter interface {
	Format(weather domain.ExchangeRate) string
}

type exchangeRate struct {
	pairs     []domain.CurrencyPair
	fetcher   ExchangeRateFetcher
	formatter ExchangeRateFormatter
}

func NewExchangeRate(
	pairs []domain.CurrencyPair,
	fetcher ExchangeRateFetcher,
	formatter ExchangeRateFormatter,
) *exchangeRate {
	return &exchangeRate{
		pairs:     pairs,
		fetcher:   fetcher,
		formatter: formatter,
	}
}

func (e *exchangeRate) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	for _, pair := range e.pairs {
		rate, err := e.fetcher.FetchLatestRate(ctx, pair)
		if err != nil {
			return "", fmt.Errorf("fetching latest exchange rate for pair %s: %v", pair, err)
		}

		sb.WriteString(e.formatter.Format(*rate))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

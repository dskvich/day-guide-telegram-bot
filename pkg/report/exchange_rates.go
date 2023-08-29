package report

import (
	"context"
	"fmt"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRatesFetcher interface {
	FetchLatest(context.Context) (*domain.USDExchangeRates, error)
}

type ExchangeRateFormatter interface {
	Format(weather domain.USDExchangeRates) string
}

type exchangeRates struct {
	fetcher   ExchangeRatesFetcher
	formatter ExchangeRateFormatter
}

func NewExchangeRates(
	fetcher ExchangeRatesFetcher,
	formatter ExchangeRateFormatter,
) *exchangeRates {
	return &exchangeRates{
		fetcher:   fetcher,
		formatter: formatter,
	}
}

func (e *exchangeRates) Generate(ctx context.Context) (string, error) {
	rates, err := e.fetcher.FetchLatest(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching latest USD exchange rates: %v", err)
	}

	return e.formatter.Format(*rates), nil
}

package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const exchangeRateAnalysisSuffix = `
Предоставь смешной обзор длинной 250-350 символов на представленные курсы валют. 
Используй эмоджи, черный юмор без границ, популярные смешные высказывания и мемы 
из русскоязычного сегмента интернета.
`

type ExchangeRateFetcher interface {
	FetchLatestRate(context.Context, domain.CurrencyPair) (*domain.ExchangeRate, error)
}

type ExchangeRateFormatter interface {
	Format(weather domain.ExchangeRate) string
}

type ExchangeRateAssistant interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type exchangeRate struct {
	pairs     []domain.CurrencyPair
	fetcher   ExchangeRateFetcher
	formatter ExchangeRateFormatter
	assistant ExchangeRateAssistant
}

func NewExchangeRate(
	pairs []domain.CurrencyPair,
	fetcher ExchangeRateFetcher,
	formatter ExchangeRateFormatter,
	assistant ExchangeRateAssistant,
) *exchangeRate {
	return &exchangeRate{
		pairs:     pairs,
		fetcher:   fetcher,
		formatter: formatter,
		assistant: assistant,
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

	resp, err := e.assistant.GetResponse(ctx, sb.String()+exchangeRateAnalysisSuffix)
	if err != nil {
		return "", fmt.Errorf("generating analysis part: %v", err)
	}

	sb.WriteString(resp)
	return sb.String(), nil
}

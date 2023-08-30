package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const exchangeRateAnalysisSuffix = "Предоставь смешной обзор на курс валют. Используй эмоджи и черный юмор без границ."

type ExchangeRateFetcher interface {
	FetchLatestRate(context.Context, domain.CurrencyPair) (*domain.ExchangeRate, error)
}

type ExchangeRateFormatter interface {
	Format(weather domain.ExchangeRate) string
}

type ExchangeRateAssistant interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type exchangeRates struct {
	pairs     []domain.CurrencyPair
	fetcher   ExchangeRateFetcher
	formatter ExchangeRateFormatter
	assistant ExchangeRateAssistant
}

func NewExchangeRates(
	pairs []domain.CurrencyPair,
	fetcher ExchangeRateFetcher,
	formatter ExchangeRateFormatter,
	assistant ExchangeRateAssistant,
) *exchangeRates {
	return &exchangeRates{
		pairs:     pairs,
		fetcher:   fetcher,
		formatter: formatter,
		assistant: assistant,
	}
}

func (e *exchangeRates) Generate(ctx context.Context) (string, error) {
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

package report

import (
	"context"
	"fmt"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const quoteAnalysisSuffix = `
Переведи цитату на русский язык и предоставь смешной анализ длинной 200-300 символов.
Добавь познавательные, исторические факты и заряд энергией.
Используй эмоджи.
`

type QuoteFetcher interface {
	FetchLatestQuote(context.Context) (*domain.Quote, error)
}

type QuoteFormatter interface {
	Format(domain.Quote) string
}

type QuoteAssistant interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type quote struct {
	fetcher   QuoteFetcher
	formatter QuoteFormatter
	assistant QuoteAssistant
}

func NewQuote(
	fetcher QuoteFetcher,
	formatter QuoteFormatter,
	assistant QuoteAssistant,
) *quote {
	return &quote{
		fetcher:   fetcher,
		formatter: formatter,
		assistant: assistant,
	}
}

func (q *quote) Generate(ctx context.Context) (string, error) {
	quote, err := q.fetcher.FetchLatestQuote(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching quote: %v", err)
	}

	quoteString := q.formatter.Format(*quote)

	/*resp, err := q.assistant.GetResponse(ctx, quoteString+quoteAnalysisSuffix)
	if err != nil {
		return "", fmt.Errorf("generating analysis part: %v", err)
	}

	return resp, nil*/
	return quoteString, nil
}

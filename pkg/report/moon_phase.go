package report

import (
	"context"
	"fmt"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const moonPhaseMessageSetupPrompt = `
Приведи по одному действию которое рекомендуется и не рекомендуется делать в этот день. Коротко.
`

type MoonPhaseFetcher interface {
	FetchLatestPhase(context.Context) (*domain.MoonPhase, error)
}

type MoonPhaseFormatter interface {
	Format(domain.MoonPhase) string
}

type moonPhase struct {
	fetcher   MoonPhaseFetcher
	formatter MoonPhaseFormatter
}

func NewMoonPhase(
	fetcher MoonPhaseFetcher,
	formatter MoonPhaseFormatter,
) *moonPhase {
	return &moonPhase{
		fetcher:   fetcher,
		formatter: formatter,
	}
}

func (m *moonPhase) Generate(ctx context.Context) (string, error) {
	phase, err := m.fetcher.FetchLatestPhase(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching latest moon phase: %v", err)
	}

	return m.formatter.Format(*phase), nil
}

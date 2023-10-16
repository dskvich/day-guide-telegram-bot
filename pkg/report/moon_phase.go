package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const moonPhaseAnalysisSuffix = `
Предоставь смешной обзор длинной 200-300 символов на представленные данные.
Добавь как называют такую луну.
Добавь познавательные и исторические факты.
`

type MoonPhaseFetcher interface {
	FetchLatestPhase(context.Context) (*domain.MoonPhase, error)
}

type MoonPhaseFormatter interface {
	Format(domain.MoonPhase) string
}

type MoonPhaseAssistant interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type moonPhase struct {
	fetcher   MoonPhaseFetcher
	formatter MoonPhaseFormatter
	assistant MoonPhaseAssistant
}

func NewMoonPhase(
	fetcher MoonPhaseFetcher,
	formatter MoonPhaseFormatter,
	assistant MoonPhaseAssistant,
) *moonPhase {
	return &moonPhase{
		fetcher:   fetcher,
		formatter: formatter,
		assistant: assistant,
	}
}

func (m *moonPhase) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	phase, err := m.fetcher.FetchLatestPhase(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching latest moon phase: %v", err)
	}

	sb.WriteString(m.formatter.Format(*phase))
	/*sb.WriteString("\n")

	resp, err := m.assistant.GetResponse(ctx, sb.String()+moonPhaseAnalysisSuffix)
	if err != nil {
		return "", fmt.Errorf("generating analysis part: %v", err)
	}

	sb.WriteString(resp)*/

	return sb.String(), nil
}

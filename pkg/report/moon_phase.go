package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const moonPhaseAnalysisSuffix = `Предоставь смешной обзор длинной 300-400 символов на представленные данные.
Добавь как называют такую луну.
Добавь познавательные и исторические факты.`

type MoonPhaseFetcher interface {
	FetchLatestPhase(context.Context) (*domain.MoonPhase, error)
}

type MoonPhaseFormatter interface {
	Format(domain.MoonPhase) string
}

type MoonPhaseAssistant interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type moonPhases struct {
	fetcher   MoonPhaseFetcher
	formatter MoonPhaseFormatter
	assistant MoonPhaseAssistant
}

func NewMoonPhases(
	fetcher MoonPhaseFetcher,
	formatter MoonPhaseFormatter,
	assistant MoonPhaseAssistant,
) *moonPhases {
	return &moonPhases{
		fetcher:   fetcher,
		formatter: formatter,
		assistant: assistant,
	}
}

func (e *moonPhases) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	phase, err := e.fetcher.FetchLatestPhase(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching latest moon phase: %v", err)
	}

	sb.WriteString(e.formatter.Format(*phase))
	sb.WriteString("\n")

	resp, err := e.assistant.GetResponse(ctx, sb.String()+moonPhaseAnalysisSuffix)
	if err != nil {
		return "", fmt.Errorf("generating analysis part: %v", err)
	}

	sb.WriteString(resp)
	return sb.String(), nil
}

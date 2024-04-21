package report

import (
	"context"
	"fmt"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const moonPhaseMessageSetupPrompt = `
Расскажи хокку про эту луну, используй пацанские цытаты.
`

type MoonPhaseFetcher interface {
	FetchLatestPhase(context.Context) (*domain.MoonPhase, error)
}

type MoonPhaseAIResponseGenerator interface {
	GenerateTextResponse(task, text string) (string, error)
}

type MoonPhaseFormatter interface {
	Format(domain.MoonPhase) string
}

type moonPhase struct {
	fetcher     MoonPhaseFetcher
	formatter   MoonPhaseFormatter
	aiGenerator MoonPhaseAIResponseGenerator
}

func NewMoonPhase(
	fetcher MoonPhaseFetcher,
	formatter MoonPhaseFormatter,
	aiGenerator MoonPhaseAIResponseGenerator,
) *moonPhase {
	return &moonPhase{
		fetcher:     fetcher,
		formatter:   formatter,
		aiGenerator: aiGenerator,
	}
}

func (m *moonPhase) Generate(ctx context.Context) (string, error) {
	phase, err := m.fetcher.FetchLatestPhase(ctx)
	if err != nil {
		return "", fmt.Errorf("fetching latest moon phase: %v", err)
	}

	phaseStr := m.formatter.Format(*phase)

	generatedStr, err := m.aiGenerator.GenerateTextResponse(moonPhaseMessageSetupPrompt, phaseStr)
	if err != nil {
		return "", fmt.Errorf("generating moon phase response with AI: %v", err)
	}

	resp := phaseStr + "\n" + generatedStr
	return resp, nil
}

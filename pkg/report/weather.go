package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const weatherAnalysisQuerySuffix = `
Проанализируй текущие погодные данные, учитывая "ощущаемую" температуру в скобках, 
и представь аналитику длинной не более 500 символов с юмором.
Используй эмоджи, черный юмор без границ, популярные высказывания и мемы из русскоязычного сегмента интернета.
`

type WeatherFetcher interface {
	FetchLatestByLocation(context.Context, domain.Location) (*domain.Weather, error)
}

type WeatherFormatter interface {
	Format(weather domain.Weather) string
}

type WeatherAssistant interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type weather struct {
	locations []domain.Location
	fetcher   WeatherFetcher
	formatter WeatherFormatter
	assistant WeatherAssistant
}

func NewWeather(
	locations []domain.Location,
	fetcher WeatherFetcher,
	formatter WeatherFormatter,
	assistant WeatherAssistant,
) *weather {
	return &weather{
		locations: locations,
		fetcher:   fetcher,
		formatter: formatter,
		assistant: assistant,
	}
}

func (w *weather) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	for _, loc := range w.locations {
		weather, err := w.fetcher.FetchLatestByLocation(ctx, loc)
		if err != nil {
			return "", fmt.Errorf("fetching latest weather for location %s: %v", loc, err)
		}

		sb.WriteString(w.formatter.Format(*weather))
		sb.WriteString("\n")
	}

	resp, err := w.assistant.GetResponse(ctx, sb.String()+weatherAnalysisQuerySuffix)
	if err != nil {
		return "", fmt.Errorf("generating analysis part: %v", err)
	}

	sb.WriteString(resp)
	return sb.String(), nil
}

package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const weatherAnalysisQuerySuffix = `Проанализируй текущие погодные данные, учитывая "ощущаемую" температуру в скобках, 
и представь аналитику с юмором. Используй эмоджи и черный юмор без границ.`

type WeatherFetcher interface {
	FetchLatestByLocation(context.Context, domain.Location) (*domain.Weather, error)
}

type TableFormatter interface {
	Format(weather domain.Weather) string
}

type GPTProvider interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type weather struct {
	locations      []domain.Location
	weatherFetcher WeatherFetcher
	tableFormatter TableFormatter
	gptProvider    GPTProvider
}

func NewWeather(
	locations []domain.Location,
	weatherFetcher WeatherFetcher,
	tableFormatter TableFormatter,
	gptProvider GPTProvider,
) *weather {
	return &weather{
		locations:      locations,
		weatherFetcher: weatherFetcher,
		tableFormatter: tableFormatter,
		gptProvider:    gptProvider,
	}
}

func (r *weather) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	for _, loc := range r.locations {
		weather, err := r.weatherFetcher.FetchLatestByLocation(ctx, loc)
		if err != nil {
			return "", fmt.Errorf("fetching latest weather for location %s: %w", loc, err)
		}

		sb.WriteString(r.tableFormatter.Format(*weather))
		sb.WriteString("\n")
	}

	resp, err := r.gptProvider.GetResponse(ctx, sb.String()+weatherAnalysisQuerySuffix)
	if err != nil {
		return "", fmt.Errorf("generating analysis part: %w", err)
	}

	sb.WriteString(resp)
	return sb.String(), nil
}

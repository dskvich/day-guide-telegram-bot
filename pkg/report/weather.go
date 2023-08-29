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

type WeatherFormatter interface {
	Format(weather domain.Weather) string
}

type GPTProvider interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

type weather struct {
	locations   []domain.Location
	fetcher     WeatherFetcher
	formatter   WeatherFormatter
	gptProvider GPTProvider
}

func NewWeather(
	locations []domain.Location,
	fetcher WeatherFetcher,
	formatter WeatherFormatter,
	gptProvider GPTProvider,
) *weather {
	return &weather{
		locations:   locations,
		fetcher:     fetcher,
		formatter:   formatter,
		gptProvider: gptProvider,
	}
}

func (r *weather) Generate(ctx context.Context) (string, error) {
	var sb strings.Builder
	for _, loc := range r.locations {
		weather, err := r.fetcher.FetchLatestByLocation(ctx, loc)
		if err != nil {
			return "", fmt.Errorf("fetching latest weather for location %s: %v", loc, err)
		}

		sb.WriteString(r.formatter.Format(*weather))
		sb.WriteString("\n")
	}

	resp, err := r.gptProvider.GetResponse(ctx, sb.String()+weatherAnalysisQuerySuffix)
	if err != nil {
		return "", fmt.Errorf("generating analysis part: %v", err)
	}

	sb.WriteString(resp)
	return sb.String(), nil
}

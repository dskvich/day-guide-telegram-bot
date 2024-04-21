package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type WeatherFetcher interface {
	FetchLatestByLocation(context.Context, domain.Location) (*domain.Weather, error)
}

type WeatherFormatter interface {
	Format(weather domain.Weather) string
}

type weather struct {
	locations []domain.Location
	fetcher   WeatherFetcher
	formatter WeatherFormatter
}

func NewWeather(
	locations []domain.Location,
	fetcher WeatherFetcher,
	formatter WeatherFormatter,
) *weather {
	return &weather{
		locations: locations,
		fetcher:   fetcher,
		formatter: formatter,
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

	return sb.String(), nil
}

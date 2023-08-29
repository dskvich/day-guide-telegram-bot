package weather

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Provider interface {
	FetchCurrent(context.Context, domain.Location) (*domain.Weather, error)
}

type Saver interface {
	Save(context.Context, *domain.Weather) error
}

type loaderService struct {
	provider        Provider
	saver           Saver
	targetLocations []domain.Location
	poolInterval    time.Duration
}

func NewLoaderService(
	provider Provider,
	saver Saver,
	targetLocations []domain.Location,
	poolInterval time.Duration,
) (*loaderService, error) {
	return &loaderService{
		provider:        provider,
		saver:           saver,
		targetLocations: targetLocations,
		poolInterval:    poolInterval,
	}, nil
}

func (_ *loaderService) Name() string { return "weather loader" }

func (s *loaderService) Run(ctx context.Context) error {
	slog.Info("starting weather loader service", "interval", s.poolInterval.String())
	defer slog.Info("stopped weather loader service")

	for {
		if err := s.load(ctx); err != nil {
			slog.Error("weather loader pass failed", logger.Err(err))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.poolInterval):
			continue
		}
	}
}

func (s *loaderService) load(ctx context.Context) error {
	slog.Info("starting weather loader pass")
	startAt := time.Now()

	for _, loc := range s.targetLocations {
		if err := s.fetchAndSave(ctx, loc); err != nil {
			slog.Error("processing location", "location", loc, logger.Err(err))
			continue
		}
	}

	slog.Info("completed weather loader pass", "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

func (s *loaderService) fetchAndSave(ctx context.Context, location domain.Location) error {
	weather, err := s.provider.FetchCurrent(ctx, location)
	if err != nil {
		return fmt.Errorf("fetching weather for location %s: %w", location, err)
	}

	if err := s.saver.Save(ctx, weather); err != nil {
		return fmt.Errorf("saving weather for location %s: %w", location, err)
	}

	return nil
}

package weather

import (
	"context"
	"log/slog"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Provider interface {
	GetCurrentWeather(city domain.City) (*domain.Weather, error)
}

type Repository interface {
	CreateNew(ctx context.Context, weather *domain.Weather) error
}

type loaderService struct {
	provider       Provider
	repo           Repository
	updateInterval time.Duration
}

func NewLoaderService(provider Provider, repo Repository) (*loaderService, error) {
	return &loaderService{
		provider:       provider,
		repo:           repo,
		updateInterval: 30 * time.Minute,
	}, nil
}

func (_ *loaderService) Name() string { return "weather loader" }

func (s *loaderService) Run(ctx context.Context) error {
	slog.Info("starting weather loader service", "update_interval", s.updateInterval.String())
	defer slog.Info("stopped weather loader service")

	if err := s.load(ctx); err != nil {
		slog.Error("weather loader pass failed", logger.Err(err))
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.updateInterval):
			continue
		}
	}
}

func (s *loaderService) load(ctx context.Context) error {
	slog.Info("starting weather loader pass")
	startAt := time.Now()

	cities := []domain.City{
		domain.SaintPetersburg,
		domain.Pitkyaranta,
		domain.Antalya,
	}

	for _, city := range cities {
		weather, err := s.provider.GetCurrentWeather(city)
		if err != nil {
			return err
		}

		if err := s.repo.CreateNew(ctx, weather); err != nil {
			return err
		}
	}

	slog.Info("completed weather loader pass", "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

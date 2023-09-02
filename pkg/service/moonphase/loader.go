package moonphase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Fetcher interface {
	FetchCurrent(context.Context) (*domain.MoonPhase, error)
}

type Saver interface {
	Save(context.Context, *domain.MoonPhase) error
}

type loaderService struct {
	fetcher      Fetcher
	saver        Saver
	poolInterval time.Duration
}

func NewLoaderService(
	fetcher Fetcher,
	saver Saver,
	poolInterval time.Duration,
) (*loaderService, error) {
	return &loaderService{
		fetcher:      fetcher,
		saver:        saver,
		poolInterval: poolInterval,
	}, nil
}

func (_ *loaderService) Name() string { return "moon phase loader" }

func (s *loaderService) Run(ctx context.Context) error {
	slog.Info("starting moon phase loader service", "interval", s.poolInterval.String())
	defer slog.Info("stopped moon phase loader service")

	for {
		if err := s.load(ctx); err != nil {
			slog.Error("moon phase loader pass failed", logger.Err(err))
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
	slog.Info("starting moon phase loader pass")
	startAt := time.Now()

	if err := s.fetchAndSave(ctx); err != nil {
		return err
	}

	slog.Info("completed moon phase loader pass", "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

func (s *loaderService) fetchAndSave(ctx context.Context) error {
	moonPhase, err := s.fetcher.FetchCurrent(ctx)
	if err != nil {
		return fmt.Errorf("fetching moon phase: %w", err)
	}

	if err := s.saver.Save(ctx, moonPhase); err != nil {
		return fmt.Errorf("saving moon phase: %w", err)
	}

	return nil
}

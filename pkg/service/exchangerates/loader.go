package exchangerates

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Fetcher interface {
	FetchCurrent(ctx context.Context) (*domain.USDExchangeRates, error)
}

type Saver interface {
	Save(context.Context, *domain.USDExchangeRates) error
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

func (s *loaderService) Name() string { return "exchange rates loader" }

func (s *loaderService) Run(ctx context.Context) error {
	slog.Info("starting exchange rates loader service", "interval", s.poolInterval.String())
	defer slog.Info("stopped exchange rates loader service")

	for {
		if err := s.load(ctx); err != nil {
			slog.Error("exchange rates loader pass failed", logger.Err(err))
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
	slog.Info("starting exchange rates loader pass")
	startAt := time.Now()

	rates, err := s.fetcher.FetchCurrent(ctx)
	if err != nil {
		return fmt.Errorf("fetching exchange rates: %v", err)
	}

	if err := s.saver.Save(ctx, rates); err != nil {
		return fmt.Errorf("saving exchange rates: %v", err)
	}

	slog.Info("completed exchange rates loader pass", "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

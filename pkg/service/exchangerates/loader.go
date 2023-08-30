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
	FetchCurrent(context.Context, domain.CurrencyPair) (*domain.ExchangeRate, error)
}

type Saver interface {
	Save(context.Context, *domain.ExchangeRate) error
}

type loaderService struct {
	pairs        []domain.CurrencyPair
	fetcher      Fetcher
	saver        Saver
	poolInterval time.Duration
}

func NewLoaderService(
	pairs []domain.CurrencyPair,
	fetcher Fetcher,
	saver Saver,
	poolInterval time.Duration,
) (*loaderService, error) {
	return &loaderService{
		pairs:        pairs,
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

	for _, pair := range s.pairs {
		if err := s.fetchAndSave(ctx, pair); err != nil {
			slog.Error("processing pair", "pair", pair, logger.Err(err))
			continue
		}
	}

	slog.Info("completed exchange rates loader pass", "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

func (s *loaderService) fetchAndSave(ctx context.Context, pair domain.CurrencyPair) error {
	rate, err := s.fetcher.FetchCurrent(ctx, pair)
	if err != nil {
		return fmt.Errorf("fetching exchange rate for pair %s: %w", pair, err)
	}

	if err := s.saver.Save(ctx, rate); err != nil {
		return fmt.Errorf("saving exchange rate for pair %s: %w", pair, err)
	}

	return nil
}

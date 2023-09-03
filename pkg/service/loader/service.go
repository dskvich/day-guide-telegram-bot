package loader

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Fetcher[T any] interface {
	FetchData(ctx context.Context) (T, error)
}

type FetcherOneParam[T any, P any] interface {
	FetchData(ctx context.Context, param P) (T, error)
}

type Saver[T any] interface {
	Save(ctx context.Context, data T) error
}

type service[T any, P any] struct {
	params       []P
	fetcher      interface{}
	saver        Saver[T]
	pollInterval time.Duration
	name         string
}

func NewService[T any, P any](
	name string,
	params []P,
	fetcher interface{},
	saver Saver[T],
	pollInterval time.Duration,
) (*service[T, P], error) {
	return &service[T, P]{
		name:         name,
		params:       params,
		fetcher:      fetcher,
		saver:        saver,
		pollInterval: pollInterval,
	}, nil
}

func (s *service[T, P]) Name() string { return s.name }

func (s *service[T, P]) Run(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s service", s.name), "interval", s.pollInterval.String())
	defer slog.Info(fmt.Sprintf("stopped %s service", s.name))

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		if err := s.load(ctx); err != nil {
			slog.Error(fmt.Sprintf("%s pass failed", s.name), logger.Err(err))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (s *service[T, P]) load(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s pass", s.name))
	startAt := time.Now()
	defer slog.Info(fmt.Sprintf("completed %s pass", s.name), "elapsed_time", time.Now().Sub(startAt).String())

	switch fetcher := s.fetcher.(type) {
	case Fetcher[T]:
		return s.fetchAndSave(ctx, fetcher)
	case FetcherOneParam[T, P]:
		return s.fetchAndSaveOneParam(ctx, fetcher)
	default:
		return fmt.Errorf("unsupported fetcher type")
	}

	return nil
}

func (s *service[T, P]) fetchAndSave(ctx context.Context, fetcher Fetcher[T]) error {
	data, err := fetcher.FetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetching data: %w", err)
	}
	if err := s.saver.Save(ctx, data); err != nil {
		return fmt.Errorf("saving data: %w", err)
	}
	return nil
}

func (s *service[T, P]) fetchAndSaveOneParam(ctx context.Context, fetcher FetcherOneParam[T, P]) error {
	for _, param := range s.params {
		data, err := fetcher.FetchData(ctx, param)
		if err != nil {
			slog.Error("fetching data for param %v: %w", param, err)
			continue
		}
		if err := s.saver.Save(ctx, data); err != nil {
			slog.Error("saving data for param %v: %w", param, err)
			continue
		}
	}
	return nil
}

package loader

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Fetcher[T any, P any] interface {
	FetchData(ctx context.Context, params P) (T, error)
}

type Saver[T any] interface {
	Save(ctx context.Context, data T) error
}

type service[T any, P any] struct {
	params       []P
	fetcher      Fetcher[T, P]
	saver        Saver[T]
	poolInterval time.Duration
	name         string
}

func NewService[T any, P any](
	name string,
	params []P,
	fetcher Fetcher[T, P],
	saver Saver[T],
	poolInterval time.Duration,
) (*service[T, P], error) {
	return &service[T, P]{
		name:         name,
		params:       params,
		fetcher:      fetcher,
		saver:        saver,
		poolInterval: poolInterval,
	}, nil
}

func (s *service[T, P]) Name() string { return s.name }

func (s *service[T, P]) Run(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s service", s.name), "interval", s.poolInterval.String())
	defer slog.Info(fmt.Sprintf("stopped %s service", s.name))

	for {
		if err := s.load(ctx); err != nil {
			slog.Error(fmt.Sprintf("%s pass failed", s.name), logger.Err(err))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.poolInterval):
			continue
		}
	}
}

func (s *service[T, P]) load(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s pass", s.name))
	startAt := time.Now()

	/*if len(s.params) == 0 {
		if err := s.fetchAndSave(ctx, (P)(struct{}{})); err != nil {
			slog.Error(fmt.Sprintf("processing without param"), logger.Err(err))
		}
	} else {*/
	for _, param := range s.params {
		if err := s.fetchAndSave(ctx, param); err != nil {
			slog.Error(fmt.Sprintf("processing param: %v", param), logger.Err(err))
			continue
		}
	}
	//}

	slog.Info(fmt.Sprintf("completed %s pass", s.name), "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

func (s *service[T, P]) fetchAndSave(ctx context.Context, param P) error {
	data, err := s.fetcher.FetchData(ctx, param)
	if err != nil {
		return fmt.Errorf("fetching data for param %v: %w", param, err)
	}

	if err := s.saver.Save(ctx, data); err != nil {
		return fmt.Errorf("saving data for param %v: %w", param, err)
	}

	return nil
}

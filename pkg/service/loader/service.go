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

func (svc *service[T, P]) Name() string { return svc.name }

func (svc *service[T, P]) Run(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s service", svc.name), "interval", svc.pollInterval.String())
	defer slog.Info(fmt.Sprintf("stopped %s service", svc.name))

	ticker := time.NewTicker(svc.pollInterval)
	defer ticker.Stop()

	for {
		if err := svc.load(ctx); err != nil {
			slog.Error(fmt.Sprintf("%s pass failed", svc.name), logger.Err(err))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func (svc *service[T, P]) load(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s pass", svc.name))
	startAt := time.Now()
	defer slog.Info(fmt.Sprintf("completed %s pass", svc.name), "elapsed_time", time.Now().Sub(startAt).String())

	switch fetcher := svc.fetcher.(type) {
	case Fetcher[T]:
		return svc.fetchAndSave(ctx, fetcher)
	case FetcherOneParam[T, P]:
		return svc.fetchAndSaveOneParam(ctx, fetcher)
	default:
		return fmt.Errorf("unsupported fetcher type")
	}

	return nil
}

func (svc *service[T, P]) fetchAndSave(ctx context.Context, fetcher Fetcher[T]) error {
	data, err := fetcher.FetchData(ctx)
	if err != nil {
		return fmt.Errorf("fetching data: %w", err)
	}
	if err := svc.saver.Save(ctx, data); err != nil {
		return fmt.Errorf("saving data: %w", err)
	}
	return nil
}

func (svc *service[T, P]) fetchAndSaveOneParam(ctx context.Context, fetcher FetcherOneParam[T, P]) error {
	for _, param := range svc.params {
		data, err := fetcher.FetchData(ctx, param)
		if err != nil {
			slog.Error("fetching data for param %v: %w", param, err)
			continue
		}
		if err := svc.saver.Save(ctx, data); err != nil {
			slog.Error("saving data for param %v: %w", param, err)
			continue
		}
	}
	return nil
}

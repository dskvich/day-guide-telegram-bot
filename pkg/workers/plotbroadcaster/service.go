package plotbroadcaster

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type ChatFetcher interface {
	GetIDs(ctx context.Context) ([]int64, error)
}

type ReportGenerator interface {
	Generate(ctx context.Context, pair domain.CurrencyPair) ([]byte, string, error)
}

type service struct {
	name            string
	cron            string
	chatFetcher     ChatFetcher
	reportGenerator ReportGenerator
	outCh           chan<- domain.Message
	pairs           []domain.CurrencyPair
}

func NewService(
	name, cron string,
	chatFetcher ChatFetcher,
	reportGenerator ReportGenerator,
	outCh chan<- domain.Message,
	pairs []domain.CurrencyPair,
) (*service, error) {
	return &service{
		name:            name,
		cron:            cron,
		chatFetcher:     chatFetcher,
		reportGenerator: reportGenerator,
		outCh:           outCh,
		pairs:           pairs,
	}, nil
}

func (svc *service) Name() string { return svc.name }

func (svc *service) Start(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s service", svc.name), "cron", svc.cron)
	defer slog.Info(fmt.Sprintf("stopped %s service", svc.name))

	c := cron.New()
	defer c.Stop()

	job := func() {
		if err := svc.broadcast(ctx); err != nil {
			slog.Error(fmt.Sprintf("%s pass failed", svc.name), logger.Err(err))
		}
	}

	if _, err := c.AddFunc(svc.cron, job); err != nil {
		slog.Error("failed to add cron job", "name", svc.name, logger.Err(err))
		return err
	}

	c.Start()
	<-ctx.Done()

	return nil
}

func (svc *service) broadcast(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s pass", svc.name))
	startAt := time.Now()

	chatIDs, err := svc.chatFetcher.GetIDs(ctx)
	if err != nil {
		return fmt.Errorf("fetching chatIDs for broadcasting: %v", err)
	}

	for _, pair := range svc.pairs {
		imageBytes, caption, err := svc.reportGenerator.Generate(context.TODO(), pair)
		if err != nil {
			for _, id := range chatIDs {
				svc.outCh <- &domain.TextMessage{
					ChatID:  id,
					Content: fmt.Sprintf("``` Failed to generate exchange rate report image for pair %s: %v ```", pair, err),
				}
			}
			continue
		}

		for _, id := range chatIDs {
			svc.outCh <- &domain.ImageMessage{
				ChatID:  id,
				Content: imageBytes,
				Caption: caption,
			}
		}
	}

	slog.Info(fmt.Sprintf("completed %s pass", svc.name), "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

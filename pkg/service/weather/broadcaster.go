package weather

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
	Generate(ctx context.Context) (string, error)
}

type broadcasterService struct {
	chatFetcher     ChatFetcher
	reportGenerator ReportGenerator
	outCh           chan<- domain.Message
}

func NewBroadcasterService(
	chatFetcher ChatFetcher,
	reportGenerator ReportGenerator,
	outCh chan<- domain.Message,
) (*broadcasterService, error) {
	return &broadcasterService{
		chatFetcher:     chatFetcher,
		reportGenerator: reportGenerator,
		outCh:           outCh,
	}, nil
}

func (_ *broadcasterService) Name() string { return "weather broadcaster" }

func (b *broadcasterService) Run(ctx context.Context) error {
	slog.Info("starting weather broadcaster service")
	defer slog.Info("stopped weather broadcaster service")

	c := cron.New()
	defer c.Stop()

	job := func() {
		if err := b.broadcast(ctx); err != nil {
			slog.Error("weather broadcaster pass failed", logger.Err(err))
		}
	}

	if _, err := c.AddFunc("0 6,10,15 * * *", job); err != nil {
		slog.Error("Failed to add cron job", logger.Err(err))
		return err
	}

	c.Start()
	<-ctx.Done()

	return nil
}

func (b *broadcasterService) broadcast(ctx context.Context) error {
	slog.Info("starting weather broadcaster pass")
	startAt := time.Now()
	defer slog.Info("completed weather broadcaster pass", "elapsed_time", time.Now().Sub(startAt).String())

	chatIDs, err := b.chatFetcher.GetIDs(ctx)
	if err != nil {
		return fmt.Errorf("fetching chatIDs for broadcasting: %w", err)
	}

	report, err := b.reportGenerator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	for _, id := range chatIDs {
		b.outCh <- &domain.TextMessage{
			ChatID:  id,
			Content: report,
		}
	}

	return nil
}

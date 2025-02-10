package workers

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

type broadcaster struct {
	name            string
	cron            string
	chatFetcher     ChatFetcher
	reportGenerator ReportGenerator
	outCh           chan<- domain.Message
}

func NewBroadcaster(
	name, cron string,
	chatFetcher ChatFetcher,
	reportGenerator ReportGenerator,
	outCh chan<- domain.Message,
) (*broadcaster, error) {
	return &broadcaster{
		name:            name,
		cron:            cron,
		chatFetcher:     chatFetcher,
		reportGenerator: reportGenerator,
		outCh:           outCh,
	}, nil
}

func (b *broadcaster) Name() string { return b.name }

func (b *broadcaster) Start(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s broadcaster", b.name), "cron", b.cron)
	defer slog.Info(fmt.Sprintf("stopped %s broadcaster", b.name))

	c := cron.New()
	defer c.Stop()

	job := func() {
		if err := b.broadcast(ctx); err != nil {
			slog.Error(fmt.Sprintf("%s pass failed", b.name), logger.Err(err))
		}
	}

	if _, err := c.AddFunc(b.cron, job); err != nil {
		slog.Error("failed to add cron job", "name", b.name, logger.Err(err))
		return err
	}

	c.Start()
	<-ctx.Done()

	return nil
}

func (b *broadcaster) broadcast(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s pass", b.name))
	startAt := time.Now()

	chatIDs, err := b.chatFetcher.GetIDs(ctx)
	if err != nil {
		return fmt.Errorf("fetching chatIDs for broadcasting: %v", err)
	}

	report, err := b.reportGenerator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("generating report: %v", err)
	}

	for _, id := range chatIDs {
		b.outCh <- &domain.TextMessage{
			ChatID:  id,
			Content: report,
		}
	}

	slog.Info(fmt.Sprintf("completed %s pass", b.name), "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

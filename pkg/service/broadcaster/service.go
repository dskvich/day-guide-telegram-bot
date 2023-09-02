package broadcaster

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

type service struct {
	name            string
	cron            string
	chatFetcher     ChatFetcher
	reportGenerator ReportGenerator
	outCh           chan<- domain.Message
}

func NewService(
	name, cron string,
	chatFetcher ChatFetcher,
	reportGenerator ReportGenerator,
	outCh chan<- domain.Message,
) (*service, error) {
	return &service{
		name:            name,
		cron:            cron,
		chatFetcher:     chatFetcher,
		reportGenerator: reportGenerator,
		outCh:           outCh,
	}, nil
}

func (s *service) Name() string { return s.name }

func (s *service) Run(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s service", s.name), "cron", s.cron)
	defer slog.Info(fmt.Sprintf("stopped %s service", s.name))

	c := cron.New()
	defer c.Stop()

	job := func() {
		if err := s.broadcast(ctx); err != nil {
			slog.Error(fmt.Sprintf("%s pass failed", s.name), logger.Err(err))
		}
	}

	if _, err := c.AddFunc(s.cron, job); err != nil {
		slog.Error("failed to add cron job", "name", s.name, logger.Err(err))
		return err
	}

	c.Start()
	<-ctx.Done()

	return nil
}

func (s *service) broadcast(ctx context.Context) error {
	slog.Info(fmt.Sprintf("starting %s pass", s.name))
	startAt := time.Now()

	chatIDs, err := s.chatFetcher.GetIDs(ctx)
	if err != nil {
		return fmt.Errorf("fetching chatIDs for broadcasting: %v", err)
	}

	report, err := s.reportGenerator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("generating report: %v", err)
	}

	for _, id := range chatIDs {
		s.outCh <- &domain.TextMessage{
			ChatID:  id,
			Content: report,
		}
	}

	slog.Info(fmt.Sprintf("completed %s pass", s.name), "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

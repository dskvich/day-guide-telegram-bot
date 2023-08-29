package weather

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Fetcher interface {
	FetchLatestByLocation(context.Context, domain.Location) (*domain.Weather, error)
}

type TableFormatter interface {
	Format(weather domain.Weather) string
}

type GptProvider interface {
	GetResponse(ctx context.Context, prompt string) (string, error)
}

const weatherAnalysisQuery = `Проанализируй текущие погодные данные, учитывая "ощущаемую" температуру в скобках, 
и представь аналитику с юмором. Используй эмоджи и черный юмор без границ.`

type broadcasterService struct {
	fetcher        Fetcher
	locations      []domain.Location
	tableFormatter TableFormatter
	outCh          chan<- string
	gptProvider    GptProvider
}

func NewBroadcasterService(
	fetcher Fetcher,
	locations []domain.Location,
	tableFormatter TableFormatter,
	outCh chan<- string,
	gptProvider GptProvider,
) (*broadcasterService, error) {
	return &broadcasterService{
		fetcher:        fetcher,
		locations:      locations,
		tableFormatter: tableFormatter,
		outCh:          outCh,
		gptProvider:    gptProvider,
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

	var sb strings.Builder
	for _, loc := range b.locations {
		weather, err := b.fetcher.FetchLatestByLocation(ctx, loc)
		if err != nil {
			return fmt.Errorf("fetching weather for location %s: %w", loc, err)
		}

		sb.WriteString(b.tableFormatter.Format(*weather))
		sb.WriteString("\n")
	}

	resp, err := b.gptProvider.GetResponse(ctx, sb.String()+weatherAnalysisQuery)
	if err != nil {
		return fmt.Errorf("generation question: %w", err)
	}

	sb.WriteString(resp)

	b.outCh <- sb.String()

	slog.Info("completed weather broadcaster pass", "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}

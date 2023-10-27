package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type QuoteReportGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type quote struct {
	reportGenerator QuoteReportGenerator
	outCh           chan<- domain.Message
}

func NewQuote(
	reportGenerator WeatherReportGenerator,
	outCh chan<- domain.Message,
) *quote {
	return &quote{
		reportGenerator: reportGenerator,
		outCh:           outCh,
	}
}

func (q *quote) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/quote")
}

func (q *quote) Handle(update *tgbotapi.Update) {
	response, err := q.reportGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate quote report: %v", err)
	}

	q.outCh <- &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

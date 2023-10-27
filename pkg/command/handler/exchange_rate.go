package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRateReportGenerator interface {
	Generate(ctx context.Context, pair domain.CurrencyPair) ([]byte, string, error)
}

type exchangeRate struct {
	reportGenerator ExchangeRateReportGenerator
	pairs           []domain.CurrencyPair
	outCh           chan<- domain.Message
}

func NewExchangeRate(
	reportGenerator ExchangeRateReportGenerator,
	pairs []domain.CurrencyPair,
	outCh chan<- domain.Message,
) *exchangeRate {
	return &exchangeRate{
		reportGenerator: reportGenerator,
		pairs:           pairs,
		outCh:           outCh,
	}
}

func (e *exchangeRate) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/rate")
}

func (e *exchangeRate) Handle(update *tgbotapi.Update) {
	for _, pair := range e.pairs {
		imageBytes, caption, err := e.reportGenerator.Generate(context.TODO(), pair)
		if err != nil {
			e.outCh <- &domain.TextMessage{
				ChatID:           update.Message.Chat.ID,
				ReplyToMessageID: update.Message.MessageID,
				Content:          fmt.Sprintf("``` Failed to generate exchange rate report image for pair %s: %v ```", pair, err),
			}
			continue
		}

		e.outCh <- &domain.ImageMessage{
			ChatID:           update.Message.Chat.ID,
			ReplyToMessageID: update.Message.MessageID,
			Content:          imageBytes,
			Caption:          caption,
		}
	}
}

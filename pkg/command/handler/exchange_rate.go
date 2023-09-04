package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRateReportGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type exchangeRate struct {
	reportGenerator WeatherReportGenerator
}

func NewExchangeRate(reportGenerator WeatherReportGenerator) *exchangeRate {
	return &exchangeRate{reportGenerator: reportGenerator}
}

func (e *exchangeRate) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/rate")
}

func (e *exchangeRate) Handle(update *tgbotapi.Update) domain.Message {
	response, err := e.reportGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate exchange rate report: %v", err)
	}
	return &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

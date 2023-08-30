package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type WeatherReportGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type weather struct {
	reportGenerator WeatherReportGenerator
}

func NewWeather(reportGenerator WeatherReportGenerator) *weather {
	return &weather{reportGenerator: reportGenerator}
}

func (w *weather) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/weather")
}

func (w *weather) Handle(update *tgbotapi.Update) domain.Message {
	response, err := w.reportGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate weather report: %v", err)
	}
	return &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

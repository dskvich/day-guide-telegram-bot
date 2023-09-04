package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type MoonPhaseReportGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type moonPhase struct {
	reportGenerator WeatherReportGenerator
}

func NewMoonPhase(reportGenerator WeatherReportGenerator) *moonPhase {
	return &moonPhase{reportGenerator: reportGenerator}
}

func (m *moonPhase) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/moon")
}

func (m *moonPhase) Handle(update *tgbotapi.Update) domain.Message {
	response, err := m.reportGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate moon phase report: %v", err)
	}
	return &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

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
	outCh           chan<- domain.Message
}

func NewMoonPhase(
	reportGenerator WeatherReportGenerator,
	outCh chan<- domain.Message,
) *moonPhase {
	return &moonPhase{
		reportGenerator: reportGenerator,
		outCh:           outCh,
	}
}

func (m *moonPhase) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/moon")
}

func (m *moonPhase) Handle(update *tgbotapi.Update) {
	response, err := m.reportGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate moon phase report: %v", err)
	}

	m.outCh <- &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

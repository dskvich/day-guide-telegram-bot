package command

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
	outCh           chan<- domain.Message
}

func NewWeather(
	reportGenerator WeatherReportGenerator,
	outCh chan<- domain.Message,
) *weather {
	return &weather{
		reportGenerator: reportGenerator,
		outCh:           outCh,
	}
}

func (w *weather) CanExecute(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/weather")
}

func (w *weather) Execute(update *tgbotapi.Update) {
	response, err := w.reportGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate weather report: %v", err)
	}

	w.outCh <- &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

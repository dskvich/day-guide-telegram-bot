package command

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type HolidayReportGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type holiday struct {
	reportGenerator HolidayReportGenerator
	outCh           chan<- domain.Message
}

func NewHoliday(
	reportGenerator HolidayReportGenerator,
	outCh chan<- domain.Message,
) *holiday {
	return &holiday{
		reportGenerator: reportGenerator,
		outCh:           outCh,
	}
}

func (_ *holiday) CanExecute(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/holiday")
}

func (h *holiday) Execute(update *tgbotapi.Update) {
	response, err := h.reportGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate holidays report: %v", err)
	}

	h.outCh <- &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

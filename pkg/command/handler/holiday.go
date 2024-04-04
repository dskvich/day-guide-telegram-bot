package handler

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type HolidaysGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type HistoricalEventsGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type holiday struct {
	holidaysGenerator         HolidaysGenerator
	historicalEventsGenerator HistoricalEventsGenerator
	outCh                     chan<- domain.Message
}

func NewHoliday(
	holidaysGenerator HolidaysGenerator,
	historicalEventsGenerator HistoricalEventsGenerator,
	outCh chan<- domain.Message,
) *holiday {
	return &holiday{
		holidaysGenerator:         holidaysGenerator,
		historicalEventsGenerator: historicalEventsGenerator,
		outCh:                     outCh,
	}
}

func (h *holiday) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/holidays")
}

func (h *holiday) Handle(update *tgbotapi.Update) {
	response, err := h.holidaysGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate holidays report: %v", err)
	}

	h.outCh <- &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}

	response, err = h.historicalEventsGenerator.Generate(context.TODO())
	if err != nil {
		response = fmt.Sprintf("Failed to generate historical events report: %v", err)
	}

	h.outCh <- &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          response,
	}
}

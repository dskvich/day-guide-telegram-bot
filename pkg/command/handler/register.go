package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/repository"
)

type Saver interface {
	Save(ctx context.Context, chat *domain.Chat) error
}

type register struct {
	saver Saver
	outCh chan<- domain.Message
}

func NewRegister(
	saver Saver,
	outCh chan<- domain.Message,
) *register {
	return &register{
		saver: saver,
		outCh: outCh,
	}
}

func (r *register) CanHandle(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/register")
}

func (r *register) Handle(update *tgbotapi.Update) {
	msg := "Registration completed"

	chat := &domain.Chat{
		ID:           update.Message.Chat.ID,
		RegisteredBy: update.Message.From.UserName,
	}
	if err := r.saver.Save(context.TODO(), chat); err != nil {
		slog.Error("registering a new chat", logger.Err(err))

		if errors.Is(err, repository.ErrChatAlreadyExists) {
			msg = "You have already registered"
		} else {
			msg = "Registration failed"
		}
	}

	r.outCh <- &domain.TextMessage{
		ChatID:           update.Message.Chat.ID,
		ReplyToMessageID: update.Message.MessageID,
		Content:          msg,
	}
}

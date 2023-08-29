package telegram

import (
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Bot interface {
	GetUpdates() tgbotapi.UpdatesChannel
	Send(message domain.Message) error
}

type CommandDispatcher interface {
	Dispatch(update tgbotapi.Update) domain.Message
}

type service struct {
	bot               Bot
	commandDispatcher CommandDispatcher
	messages          chan domain.Message
}

func NewService(
	bot Bot,
	commandDispatcher CommandDispatcher,
	messages chan domain.Message,
) (*service, error) {
	return &service{
		bot:               bot,
		commandDispatcher: commandDispatcher,
		messages:          messages,
	}, nil
}

func (s *service) Name() string { return "telegram bot" }

func (s *service) Run(ctx context.Context) error {
	slog.Info("starting telegram bot service")
	defer slog.Info("stopped telegram bot service")

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-s.bot.GetUpdates():
			go s.handleUpdate(update)
		case message := <-s.messages:
			go s.handleMessage(message)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *service) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		slog.Info("update message received", "chat", update.Message.Chat.ID, "user", update.Message.From, "text", update.Message.Text)
		s.messages <- s.commandDispatcher.Dispatch(update)
	}
}

func (s *service) handleMessage(message domain.Message) {
	if err := s.bot.Send(message); err != nil {
		slog.Error("sending message", logger.Err(err))
	}
}

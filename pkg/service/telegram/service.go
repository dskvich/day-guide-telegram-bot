package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Authenticator interface {
	IsAuthorized(userID int64) bool
}

type Bot interface {
	GetUpdates() tgbotapi.UpdatesChannel
	Send(message domain.Message) error
}

type CommandDispatcher interface {
	ExecuteCommands(update tgbotapi.Update)
}

type service struct {
	bot               Bot
	authenticator     Authenticator
	commandDispatcher CommandDispatcher
	messages          chan domain.Message
}

func NewService(
	bot Bot,
	authenticator Authenticator,
	commandDispatcher CommandDispatcher,
	messages chan domain.Message,
) (*service, error) {
	return &service{
		bot:               bot,
		authenticator:     authenticator,
		commandDispatcher: commandDispatcher,
		messages:          messages,
	}, nil
}

func (svc *service) Name() string { return "telegram bot" }

func (svc *service) Run(ctx context.Context) error {
	slog.Info("starting telegram bot service")
	defer slog.Info("stopped telegram bot service")

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-svc.bot.GetUpdates():
			go svc.handleUpdate(update)
		case message := <-svc.messages:
			go svc.handleMessage(message)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (svc *service) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		slog.Info("update message received", "chat", update.Message.Chat.ID, "user", update.Message.From, "text", update.Message.Text)

		if !svc.authenticator.IsAuthorized(update.Message.From.ID) {
			svc.messages <- &domain.TextMessage{
				ChatID:           update.Message.Chat.ID,
				ReplyToMessageID: update.Message.MessageID,
				Content:          fmt.Sprintf("User ID %d not authorized to use this bot.", update.Message.From.ID),
			}
			return
		}

		svc.commandDispatcher.ExecuteCommands(update)
	}
}

func (svc *service) handleMessage(message domain.Message) {
	if err := svc.bot.Send(message); err != nil {
		slog.Error("sending message", "message", message, logger.Err(err))
	}
}

package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/repository"
)

type UpdateProcessor interface {
	Process(update tgbotapi.Update)
}

type ChatRepository interface {
	GetIDs(ctx context.Context) ([]int64, error)
	CreateNew(ctx context.Context, chat *domain.Chat) error
}

type botService struct {
	bot      *tgbotapi.BotAPI
	repo     ChatRepository
	messages chan string
}

func NewBotService(botToken string, repo ChatRepository, messages chan string) (*botService, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot api: %v", err)
	}

	slog.Info("authorized on telegram", "account", bot.Self.UserName)

	return &botService{
		bot:      bot,
		repo:     repo,
		messages: messages,
	}, nil
}

func (_ *botService) Name() string { return "telegram bot" }

func (b *botService) Run(ctx context.Context) error {
	slog.Info("starting telegram bot service")
	defer slog.Info("stopped telegram bot service")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			go b.handleUpdate(update)
		case message := <-b.messages:
			go b.handleMessage(message)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (b *botService) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		slog.Info("update message received", "chat", update.Message.Chat.ID, "user", update.Message.From, "text", update.Message.Text)

		switch update.Message.Text {
		case "/register":
			msg := "Registration completed"

			chat := &domain.Chat{
				ID:           update.Message.Chat.ID,
				RegisteredBy: update.Message.From.UserName,
			}
			if err := b.repo.CreateNew(context.TODO(), chat); err != nil {
				slog.Error("registering a new chat", logger.Err(err))

				if errors.Is(err, repository.ErrChatAlreadyExists) {
					msg = "You have already registered"
				} else {
					msg = "Registration failed"
				}
			}

			b.messages <- msg
		}
	}
}

func (b *botService) handleMessage(message string) {
	chatIDs, err := b.repo.GetIDs(context.TODO())
	if err != nil {
		slog.Error("fetching chat ids from db", logger.Err(err))
	}

	for _, id := range chatIDs {
		msg := tgbotapi.NewMessage(id, message)

		if _, err := b.bot.Send(msg); err != nil {
			slog.Error("sending message", logger.Err(err))
		}
	}
}

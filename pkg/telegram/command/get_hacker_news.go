package command

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type HackerNewsService interface {
	GetNews(limit int) ([]domain.NewsItem, error)
	GetNewsAsText(limit int) (string, error)
}

type AIClient interface {
	GenerateResponse(ctx context.Context, messages []domain.GMessage) (domain.GMessage, error)
}

type TelegramClient interface {
	SendResponse(ctx context.Context, chatID int64, response string)
	SendError(ctx context.Context, chatID int64, err error)
}

type getHackerNews struct {
	service        HackerNewsService
	aiClient       AIClient
	telegramClient TelegramClient
}

func NewGetHackerNews(
	service HackerNewsService,
	aiClient AIClient,
	telegramClient TelegramClient,
) *getHackerNews {
	return &getHackerNews{
		service:        service,
		aiClient:       aiClient,
		telegramClient: telegramClient,
	}
}

func (g *getHackerNews) CanExecute(update *tgbotapi.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/news")
}

func (g *getHackerNews) Execute(update *tgbotapi.Update) {
	ctx := context.Background()
	text, err := g.service.GetNewsAsText(10)
	if err != nil {
		g.telegramClient.SendError(ctx, update.Message.Chat.ID, err)
		return
	}

	resp, err := g.aiClient.GenerateResponse(context.Background(), []domain.GMessage{
		{
			Role: "user",
			Parts: []domain.GMessagePart{
				{
					Text: "Предоставь сводку новостей на русском языке. Предоставь ссылки и рейтинг статей.",
				},
				{
					Text: text,
				},
			},
		},
	})
	if err != nil {
		g.telegramClient.SendError(ctx, update.Message.Chat.ID, err)
		return
	}

	g.telegramClient.SendResponse(ctx, update.Message.Chat.ID, resp.Parts[0].Text)
}

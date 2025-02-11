package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/render"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const maxTelegramMessageLength = 4096

type client struct {
	token   string
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

func NewClient(token string) (*client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("creating bot api: %v", err)
	}

	slog.Info("authorized on telegram", "bot", bot.Self)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return &client{
		token:   token,
		bot:     bot,
		updates: bot.GetUpdatesChan(u),
	}, nil
}

func (c *client) GetUpdates() tgbotapi.UpdatesChannel {
	return c.updates
}

func (c *client) Send(message domain.Message) error {
	if _, err := c.bot.Send(message.ToChatMessage()); err != nil {
		return fmt.Errorf("sending message: %v", err)
	}
	return nil
}

func (c *client) SendResponse(ctx context.Context, chatID int64, response string) {
	if response != "" {
		if err := c.sendText(ctx, chatID, response); err != nil {
			c.handleError(ctx, chatID, err)
		}
	}
}

func (c *client) handleError(ctx context.Context, chatID int64, err error) {
	slog.ErrorContext(ctx, "Error during sending message", logger.Err(err))

	m := tgbotapi.NewMessage(chatID, "❌ Не удалось доставить ответ")

	_, _ = c.bot.Send(m)
}

func (c *client) SendError(ctx context.Context, chatID int64, err error) {
	slog.ErrorContext(ctx, "error occurred", "chatID", chatID, logger.Err(err))

	if err := c.sendText(ctx, chatID, fmt.Sprintf("❌ %s", err.Error())); err != nil {
		c.handleError(ctx, chatID, err)
	}
}

func (c *client) sendText(ctx context.Context, chatID int64, text string) error {
	htmlText := render.ToHTML(text)

	for htmlText != "" {
		if utf8.RuneCountInString(htmlText) <= maxTelegramMessageLength {
			if err := c.send(chatID, htmlText); err != nil {
				return err
			}
			return nil
		}

		cutIndex := c.findCutIndex(htmlText, maxTelegramMessageLength)
		if err := c.send(chatID, htmlText); err != nil {
			return err
		}
		htmlText = htmlText[cutIndex:]

		// 1 message per second
		time.Sleep(time.Second)
	}

	return nil
}

func (c *client) findCutIndex(text string, maxLength int) int {
	lastPre := strings.LastIndex(text[:maxLength], "<pre>")
	lastNewline := strings.LastIndex(text[:maxLength], "\n")

	if lastPre > -1 {
		return lastPre
	}
	if lastNewline > -1 {
		return lastNewline
	}
	return maxLength
}

func (c *client) send(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true

	_, err := c.bot.Send(msg)
	return err
}

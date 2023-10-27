package domain

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type ImageMessage struct {
	ChatID           int64
	ReplyToMessageID int
	Prompt           string
	Content          []byte
	Caption          string
}

func (i *ImageMessage) ToChatMessage() tgbotapi.Chattable {
	fileBytes := tgbotapi.FileBytes{
		Bytes: i.Content,
	}
	msg := tgbotapi.NewPhoto(i.ChatID, fileBytes)

	if i.Caption != "" {
		msg.Caption = i.Caption
		msg.ParseMode = tgbotapi.ModeMarkdown
	}

	return msg
}

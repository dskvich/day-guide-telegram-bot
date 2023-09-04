package formatter

import (
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type Quote struct{}

func (_ *Quote) Format(q domain.Quote) string {
	var sb strings.Builder

	sb.WriteString(q.Quote)
	sb.WriteString(" (c)")
	sb.WriteString(q.Author)
	sb.WriteString("\n")

	return sb.String()
}

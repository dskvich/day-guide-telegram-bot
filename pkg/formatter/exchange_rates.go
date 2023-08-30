package formatter

import (
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRate struct{}

func (_ *ExchangeRate) Format(e domain.ExchangeRate) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("💲%s%s: *%.2f* \n", e.Pair.Base, e.Pair.Quote, e.Rate))

	return sb.String()
}

package formatter

import (
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRate struct{}

func (_ *ExchangeRate) Format(e domain.ExchangeRateInfo) string {
	var sb strings.Builder

	arrowIcon := ""
	percentageChange := e.PercentageChange()
	if percentageChange < 0 {
		arrowIcon = "🔽"
	} else {
		arrowIcon = "🔼"
	}

	if arrowIcon != "" {
		sb.WriteString(arrowIcon)
	}

	sb.WriteString(fmt.Sprintf(" %s/%s: *%.2f*",
		e.CurrentRate.Pair.Base,
		e.CurrentRate.Pair.Quote,
		e.CurrentRate.Rate,
	))

	if percentageChange != 0 {
		sb.WriteString(fmt.Sprintf(" %.2f%%", percentageChange))
	}

	return sb.String()
}

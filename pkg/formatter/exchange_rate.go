package formatter

import (
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRate struct{}

func (_ *ExchangeRate) Format(e domain.ExchangeRateInfo) string {
	var sb strings.Builder

	percentageChange := e.PercentageChange()
	arrowIcon := ""

	switch {
	case percentageChange > 0:
		arrowIcon = "🔺"
	case percentageChange < 0:
		arrowIcon = "🔻"
	}

	// arrow
	if arrowIcon != "" {
		sb.WriteString(arrowIcon)
	}

	// header and current rate
	sb.WriteString(fmt.Sprintf(" %s/%s:", e.CurrentRate.Pair.Base, e.CurrentRate.Pair.Quote))
	sb.WriteString(fmt.Sprintf(" *%.2f*", e.CurrentRate.Rate))

	// percentage change
	if percentageChange != 0 {
		sb.WriteString(" ")
		if percentageChange > 0 {
			sb.WriteString("+")
		}
		sb.WriteString(fmt.Sprintf("%.2f%%", percentageChange))
	}

	return sb.String()
}

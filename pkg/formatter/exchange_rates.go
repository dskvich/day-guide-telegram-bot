package formatter

import (
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchageRates struct{}

func (_ *ExchageRates) Format(r domain.USDExchangeRates) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("💵 USD Exchange Rates 💵\n"))
	sb.WriteString(fmt.Sprintf("RUB: *%.2f* \n", r.RUB))
	sb.WriteString(fmt.Sprintf("TRY: *%.2f* \n", r.TRY))

	return sb.String()
}

func truncate(input float64) float64 {
	return float64(int(input*100)) / 100.0
}

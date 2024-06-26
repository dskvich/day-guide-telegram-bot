package formatter

import (
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type MoonPhase struct{}

var phaseEmoji = map[string]string{
	"Dark Moon":       "🌑",
	"New Moon":        "🌑",
	"Waxing Crescent": "🌒",
	"1st Quarter":     "🌓",
	"Full Moon":       "🌕",
	"Waning Crescent": "🌔",
	"3rd Quarter":     "🌗",
	"Waning Gibbous":  "🌖",
	"Waxing Gibbous":  "🌔",
}

func (_ *MoonPhase) Format(m domain.MoonPhase) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s, *%d-й* лунный день\n", phaseEmoji[m.Phase], moonPhaseDescription(m.Phase), m.Age))

	return sb.String()
}

func moonPhaseDescription(phase string) string {
	switch {
	case phase == "New Moon":
		return "Новолуние"
	case phase == "Waxing Crescent":
		return "Растущий серп"
	case phase == "1st Quarter":
		return "Первая четверть"
	case phase == "Waxing Gibbous":
		return "Растущая Луна"
	case phase == "Full Moon":
		return "Полнолуние"
	case phase == "Waning Gibbous":
		return "Убывающая Луна"
	case phase == "3rd Quarter":
		return "Последняя четверть"
	case phase == "Waning Crescent":
		return "Убывающий серп"
	case phase == "Dark Moon":
		return "Тёмное новолуние"
	default:
		return fmt.Sprintf("Неизвестная фаза Луны -  %s", phase)
	}
}

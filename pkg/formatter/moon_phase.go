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

	sb.WriteString(fmt.Sprintf("%s %s, *%d-й* лунный день\n", phaseEmoji[m.Phase], moonPhaseDescription(m.Phase, m.IlluminationPrc), m.Age))

	return sb.String()
}

func moonPhaseDescription(phase string, visibility int) string {
	switch {
	case phase == "New Moon" && visibility == 0:
		return "Новолуние"
	case phase == "Waxing Crescent" && visibility > 0 && visibility < 25:
		return "Растущий серп"
	case phase == "1st Quarter" && visibility >= 25 && visibility < 50:
		return "Первая четверть"
	case phase == "Waxing Gibbous" && visibility == 50:
		return "Полумесяц"
	case phase == "Waxing Gibbous" && visibility > 50 && visibility < 75:
		return "Растущая Луна"
	case phase == "Full Moon" && (visibility >= 75 && visibility <= 100):
		return "Полнолуние"
	case phase == "Waning Gibbous" && visibility > 75 && visibility < 100:
		return "Убывающая Луна"
	case phase == "3rd Quarter" && visibility >= 50 && visibility < 75:
		return "Последняя четверть"
	case phase == "Waning Crescent" && visibility >= 25 && visibility < 50:
		return "Убывающий полумесяц"
	case phase == "Waning Crescent" && visibility > 0 && visibility < 25:
		return "Убывающий серп"
	default:
		return fmt.Sprintf("Неизвестная фаза Луны. Фаза %s, Видимость %d%%", phase, visibility)
	}
}

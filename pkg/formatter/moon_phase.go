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
	case phase == "New Moon" && (visibility == 0 || visibility == 1 || visibility == 2):
		return "Новолуние"
	case phase == "Waxing Crescent" && (visibility >= 2 && visibility <= 50):
		return "Растущий серп"
	case phase == "1st Quarter" && (visibility >= 50 && visibility <= 56):
		return "Первая четверть"
	case phase == "Waxing Gibbous" && (visibility >= 56 && visibility <= 75):
		return "Растущая Луна"
	case phase == "Full Moon" && visibility == 100:
		return "Полнолуние"
	case phase == "Waning Gibbous" && (visibility >= 53 && visibility <= 100):
		return "Убывающая Луна"
	case phase == "3rd Quarter" && (visibility >= 47 && visibility <= 53):
		return "Последняя четверть"
	case phase == "Waning Crescent" && (visibility >= 1 && visibility <= 47):
		return "Убывающий серп"
	case phase == "Dark Moon" && (visibility == 0 || visibility == 1):
		return "Тёмное новолуние"
	default:
		return fmt.Sprintf("Неизвестная фаза Луны. Фаза %s, Видимость %d%%", phase, visibility)
	}
}

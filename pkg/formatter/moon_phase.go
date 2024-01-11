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

	sb.WriteString(fmt.Sprintf("%s %s, *%d-й* лунный день\n", phaseEmoji[m.Phase], moonPhaseDescription(m.IlluminationPrc), m.Age))
	sb.WriteString(fmt.Sprintf("🌍 Расстояние до Земли: *%d км*\n", int(m.DistanceToEarth)))
	sb.WriteString(fmt.Sprintf("🌞 Расстояние до Солнца: *%d км*\n", int(m.DistanceToSun)))

	return sb.String()
}

func moonPhaseDescription(visibility int) string {
	switch {
	case visibility == 0:
		return "Новолуние"
	case visibility > 0 && visibility < 25:
		return "Растущий серп"
	case visibility >= 25 && visibility < 50:
		return "Первая четверть"
	case visibility == 50:
		return "Полумесяц"
	case visibility > 50 && visibility < 75:
		return "Растущая гиббозная Луна"
	case visibility >= 75 && visibility < 100:
		return "Полнолуние"
	case visibility > 75 && visibility < 100:
		return "Убывающая гиббозная Луна"
	case visibility >= 50 && visibility < 75:
		return "Последняя четверть"
	case visibility > 25 && visibility < 50:
		return "Убывающий полумесяц"
	case visibility > 0 && visibility < 25:
		return "Убывающий серп"
	default:
		return "Неизвестная фаза Луны"
	}
}

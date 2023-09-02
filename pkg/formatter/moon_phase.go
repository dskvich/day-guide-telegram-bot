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

	sb.WriteString(fmt.Sprintf("%s %s, *%d-й* лунный день\n", phaseEmoji[m.Phase], translateMoonPhase(m.Phase), m.Age))
	sb.WriteString(fmt.Sprintf("💡 Видимость: *%d%%*\n", m.IlluminationPrc))
	sb.WriteString(fmt.Sprintf("🌍 Расстояние до Земли: *%d км*\n", int(m.DistanceToEarth)))
	sb.WriteString(fmt.Sprintf("🌞 Расстояние до Солнца: *%d км*\n", int(m.DistanceToSun)))

	return sb.String()
}

func translateMoonPhase(phase string) string {
	switch phase {
	case "First Quarter":
		return "Первая четверть"
	case "Full":
		return "Полнолуние"
	case "Last Quarter":
		return "Последняя четверть"
	case "New":
		return "Новолуние"
	case "New Crescent":
		return "Молодая луна"
	case "Old Crescent":
		return "Старая луна"
	case "Waning Gibbous":
		return "Убывающая луна"
	case "Waxing Gibbous":
		return "Растущая луна"
	default:
		return ""
	}
}

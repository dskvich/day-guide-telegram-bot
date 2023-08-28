package formatters

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type TableFormatter struct{}

func (_ *TableFormatter) Format(w domain.Weather) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s - %s\n", weatherEmoji(w.Weather), w.Location, w.WeatherVerbose))
	sb.WriteString(fmt.Sprintf("🌡️ Температура *%.1f°C* (*%.1f°C*)\n", w.Temp, w.TempFeel))
	sb.WriteString(fmt.Sprintf("💧 Влажность *%d%%*\n", w.Humidity))
	sb.WriteString(fmt.Sprintf("🌀 Ветер *%s*\n", windDescription(w.WindSpeed, w.WindDirection)))
	sb.WriteString(fmt.Sprintf("📉 Давление *%dмм*\n", HPaToMmHg(w.Pressure)))

	return sb.String()
}

func weatherEmoji(weatherMain string) string {
	now := time.Now().Hour()

	switch weatherMain {
	case "Clear":
		if 6 <= now && now < 18 {
			return "☀️"
		} else {
			return "🌙"
		}
	case "Rain", "Drizzle":
		return "🌧️"
	case "Clouds":
		return "☁️"
	case "Fog", "Mist":
		return "🌫️"
	case "Thunderstorm":
		return "🌧️⚡"
	case "Snow":
		return "🌨️❄️"
	default:
		return "❓"
	}
}

// HPaToMmHg - converts hectopascal to mm of mercury
func HPaToMmHg(hPa int) int {
	return int(math.Round(float64(hPa) * 0.75006375541921))
}

func windDescription(windSpeed float64, windDirection string) string {
	if windSpeed == 0.0 {
		return "отсутствует"
	}
	return fmt.Sprintf("%s %.1fм/c ", windDirection, windSpeed)
}
package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type HolidaysFetcher interface {
	FetchByDate(ctx context.Context, date time.Time) ([]domain.Holiday, error)
}

type HolidaysFormatter interface {
	Format(holidays []domain.Holiday) string
}

type holiday struct {
	fetcher HolidaysFetcher
}

func NewHoliday(
	fetcher HolidaysFetcher,
) *holiday {
	return &holiday{
		fetcher: fetcher,
	}
}

func (h *holiday) Generate(ctx context.Context) (string, error) {
	now := time.Now()

	holidays, err := h.fetcher.FetchByDate(ctx, now)
	if err != nil {
		return "", fmt.Errorf("fetching holidays for date %s: %v", now, err)
	}

	if len(holidays) == 0 {
		return "Сегодня нет официальных праздников. Наслаждайтесь обычным днём!", nil
	}

	holidaysStr := joinHolidays(holidays)

	resp := fmt.Sprintf("🎉 *%s: Какие праздники отмечаем?* 🎉\n\n", formatDate(now)) + holidaysStr
	return resp, nil
}

func joinHolidays(holidays []domain.Holiday) string {
	names := make([]string, 0, len(holidays))
	for _, holiday := range holidays {
		var icons string
		for _, category := range holiday.Categories {
			icons += getEmoji(category)
		}
		names = append(names, fmt.Sprintf("%s %s", icons, holiday.Name))
	}

	return strings.Join(names, "\n")
}

func getEmoji(category string) string {
	switch category {
	case "Международные праздники":
		return "🌍"
	case "Праздники России":
		return "🇷🇺"
	case "Праздники славян":
		return "🪆"
	case "Праздники ООН":
		return "🤝"
	case "Православные праздники":
		return "✝️"
	default:
		return "❓"
	}
}

// TODO: create formatter
func formatDate(t time.Time) string {
	p := message.NewPrinter(language.Russian)
	return p.Sprintf("%d %s", t.Day(), russianMonths()[t.Month()-1])
}

func russianMonths() []string {
	return []string{
		"января", "февраля", "марта", "апреля", "мая", "июня", "июля",
		"августа", "сентября", "октября", "ноября", "декабря",
	}
}

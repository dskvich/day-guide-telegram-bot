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

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Праздники %s:\n", formatDate(now)))

	for _, holiday := range holidays {
		sb.WriteString(fmt.Sprintf("• %s\n", holiday.Name))
	}

	return sb.String(), nil
}

// TODO: create formatter
func formatDate(t time.Time) string {
	p := message.NewPrinter(language.Russian)
	return p.Sprintf("%d %s %d г.", t.Day(), russianMonths()[t.Month()-1], t.Year())
}

func russianMonths() []string {
	return []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
}

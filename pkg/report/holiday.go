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

const holidayMessageSetupPrompt = `
Создай оповещение о сегодняшних праздниках для телеграм-бота.
Включи в сообщение эмодзи, затем название праздника на русском языке (помести * с обоих сторон), и добавь пару слов от себя.
`

type HolidaysFetcher interface {
	FetchByDate(ctx context.Context, date time.Time) ([]domain.Holiday, error)
}

type HolidaysFormatter interface {
	Format(holidays []domain.Holiday) string
}

type HolidaysAIResponseGenerator interface {
	GenerateTextResponse(task, text string) (string, error)
}

type holiday struct {
	fetcher     HolidaysFetcher
	aiGenerator HolidaysAIResponseGenerator
}

func NewHoliday(
	fetcher HolidaysFetcher,
	aiGenerator HolidaysAIResponseGenerator,
) *holiday {
	return &holiday{
		fetcher:     fetcher,
		aiGenerator: aiGenerator,
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

	generatedStr, err := h.aiGenerator.GenerateTextResponse(holidayMessageSetupPrompt, holidaysStr)
	if err != nil {
		return "", fmt.Errorf("generating holidays response with AI: %v", err)
	}

	resp := fmt.Sprintf("🎉 *Праздники %s* 🎉\n\n", formatDate(now)) + generatedStr
	return resp, nil
}

func joinHolidays(holidays []domain.Holiday) string {
	names := make([]string, 0, len(holidays))
	for _, holiday := range holidays {
		names = append(names, holiday.Name)
	}

	return strings.Join(names, "\n")
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

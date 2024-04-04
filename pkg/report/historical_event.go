package report

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type historicalEvent struct {
}

func NewHistoricalEvent() *historicalEvent {
	return &historicalEvent{}
}

func (h *historicalEvent) Generate(ctx context.Context) (string, error) {
	page, err := h.fetchPage("https://kakoysegodnyaprazdnik.ru/")
	if err != nil {
		return "", fmt.Errorf("fetching holiday page: %s", err)
	}

	return h.getHistoricalEvents(page), nil
}

func (h *historicalEvent) getHistoricalEvents(doc *goquery.Document) string {
	var events []string
	doc.Find("div.event").Each(func(i int, s *goquery.Selection) {
		eventText := strings.ReplaceAll(s.Text(), "• ", "")
		if span := s.Find("span"); span.Length() > 0 {
			eventText = strings.ReplaceAll(eventText, span.Text(), "")
		}
		events = append(events, eventText)
	})
	return fmt.Sprintf("События в истории %s:\n• %s\n", h.formatDate(time.Now()), strings.Join(events, "\n• "))
}

func (h *historicalEvent) formatDate(now time.Time) string {
	p := message.NewPrinter(language.Russian)
	dateFormat := p.Sprintf("%d %%s", now.Day())

	russianMonths := []string{
		"Января", "Февраля", "Марта", "Апреля", "Мая", "Июня",
		"Июля", "Августа", "Сентября", "Октября", "Ноября", "Декабря",
	}

	return fmt.Sprintf(dateFormat, russianMonths[now.Month()-1])
}

func (h *historicalEvent) fetchPage(url string) (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var headers = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

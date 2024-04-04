package report

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type holiday struct {
}

func NewHoliday() *holiday {
	return &holiday{}
}

func (h *holiday) Generate(ctx context.Context) (string, error) {
	page, err := h.fetchPage("https://kakoysegodnyaprazdnik.ru/")
	if err != nil {
		return "", fmt.Errorf("fetching holiday page: %s", err)
	}

	return h.getHolidays(page), nil
}

func (h *holiday) getHolidays(doc *goquery.Document) string {
	var holidays []string
	doc.Find("span[itemprop='text']").Each(func(i int, s *goquery.Selection) {
		holidays = append(holidays, s.Text())
	})
	return fmt.Sprintf("Праздники %s:\n• %s\n", h.formatDate(time.Now()), strings.Join(holidays, "\n• "))
}

func (h *holiday) formatDate(now time.Time) string {
	p := message.NewPrinter(language.Russian)
	dateFormat := p.Sprintf("%d %%s", now.Day())

	russianMonths := []string{
		"Января", "Февраля", "Марта", "Апреля", "Мая", "Июня",
		"Июля", "Августа", "Сентября", "Октября", "Ноября", "Декабря",
	}

	return fmt.Sprintf(dateFormat, russianMonths[now.Month()-1])
}

func (h *holiday) fetchPage(url string) (*goquery.Document, error) {
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

	slog.Info("holidays loaded", "status", resp.Status)

	return goquery.NewDocumentFromReader(resp.Body)
}

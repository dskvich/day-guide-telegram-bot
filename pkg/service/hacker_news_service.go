package service

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/parser"
)

type HackerNewsParser interface {
	Parse(html string) ([]domain.NewsItem, error)
}

type HackerNewsService struct {
	URL    string
	Parser HackerNewsParser
}

func NewsHackerNewsService() *HackerNewsService {
	return &HackerNewsService{
		URL:    "https://news.ycombinator.com/news",
		Parser: parser.HackerNewsParser{},
	}
}

func (s HackerNewsService) GetNews(limit int) ([]domain.NewsItem, error) {
	html, err := s.fetchHTML(s.URL)
	if err != nil {
		return nil, err
	}
	items, err := s.Parser.Parse(html)
	if err != nil {
		return nil, err
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

func (s HackerNewsService) GetNewsAsText(limit int) (string, error) {
	items, err := s.GetNews(limit)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(item.ToText())
		sb.WriteString("\n" + strings.Repeat("-", 80) + "\n")
	}
	return sb.String(), nil
}

func (s *HackerNewsService) fetchHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 HTTP status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return string(body), nil
}

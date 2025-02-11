package parser

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type HackerNewsParser struct{}

func (p HackerNewsParser) Parse(html string) ([]domain.NewsItem, error) {
	// Load the HTML document with goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	var items []domain.NewsItem

	// Each submission is wrapped in a <tr> with class "athing".
	doc.Find("tr.athing").Each(func(i int, s *goquery.Selection) {
		var item domain.NewsItem

		// Get submission ID.
		if id, exists := s.Attr("id"); exists {
			item.ID = id
		}

		// Get rank (e.g., "1.").
		rankText := s.Find("td.title > span.rank").Text()
		rankText = strings.TrimSuffix(rankText, ".")
		if rank, err := strconv.Atoi(rankText); err == nil {
			item.Rank = rank
		}

		// Get title and URL.
		titleSel := s.Find("span.titleline a")
		item.Title = strings.TrimSpace(titleSel.Text())
		if url, exists := titleSel.Attr("href"); exists {
			item.URL = url
		}

		// Get the site (if available).
		item.Site = s.Find("span.sitebit a span.sitestr").Text()

		// The subtext row is the next sibling <tr>.
		subtext := s.Next().Find("td.subtext")

		// Extract score, e.g., "691 points".
		scoreText := subtext.Find("span.score").Text()
		if scoreText != "" {
			scoreParts := strings.Split(scoreText, " ")
			if len(scoreParts) > 0 {
				if score, err := strconv.Atoi(scoreParts[0]); err == nil {
					item.Score = score
				}
			}
		}

		// Extract the author.
		item.Author = subtext.Find("a.hnuser").Text()

		// Extract the age.
		item.Age = subtext.Find("span.age a").Text()

		// Extract comment count. Some submissions (like job posts) might not have comments.
		commentText := subtext.Find("a").Last().Text()
		if strings.Contains(commentText, "comment") {
			if strings.Contains(commentText, "discuss") {
				item.Comments = 0
			} else {
				parts := strings.Fields(commentText)
				if len(parts) > 0 {
					if comments, err := strconv.Atoi(parts[0]); err == nil {
						item.Comments = comments
					}
				}
			}
		}

		items = append(items, item)
	})

	return items, nil
}

func fetchHTML(url string) (string, error) {
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

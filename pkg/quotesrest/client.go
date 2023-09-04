package quotesrest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const baseURL = "http://quotes.rest/qod.json"

type client struct {
	apiKey string
	hc     *http.Client
}

func NewClient(apiKey string) *client {
	return &client{
		apiKey: apiKey,
		hc:     &http.Client{},
	}
}

func (c *client) FetchData(ctx context.Context) (*domain.Quote, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %v", err)
	}

	q := u.Query()
	q.Set("category", "inspire")
	q.Set("api_key", c.apiKey)

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %v", err)
	}

	var res quoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response body: %v", err)
	}

	if len(res.Contents.Quotes) == 0 {
		return nil, fmt.Errorf("no quotes available for today")
	}

	return &domain.Quote{
		Quote:  res.Contents.Quotes[0].Quote,
		Author: res.Contents.Quotes[0].Author,
	}, nil
}

type quoteResponse struct {
	Success struct {
		Total int `json:"total"`
	} `json:"success"`
	Contents struct {
		Quotes []struct {
			Id         string   `json:"id"`
			Quote      string   `json:"quote"`
			Length     int      `json:"length"`
			Author     string   `json:"author"`
			Language   string   `json:"language"`
			Tags       []string `json:"tags"`
			Sfw        string   `json:"sfw"`
			Permalink  string   `json:"permalink"`
			Title      string   `json:"title"`
			Category   string   `json:"category"`
			Background string   `json:"background"`
			Date       string   `json:"date"`
		} `json:"quotes"`
	} `json:"contents"`
	Copyright struct {
		Url  string `json:"url"`
		Year string `json:"year"`
	} `json:"copyright"`
}

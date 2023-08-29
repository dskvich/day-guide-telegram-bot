package openexchangerates

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const baseURL = "https://openexchangerates.org/api/latest.json"

type client struct {
	appID string
	hc    *http.Client
}

func NewClient(appID string) *client {
	return &client{
		appID: appID,
		hc:    &http.Client{},
	}
}

func (c *client) FetchCurrent(ctx context.Context) (*domain.USDExchangeRates, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %v", err)
	}

	q := u.Query()
	q.Set("app_id", c.appID)
	q.Set("symbols", "RUB,TRY")

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %v", err)
	}

	var res usdExchangeRateAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response body: %v", err)
	}

	return &domain.USDExchangeRates{
		RUB: res.Rates.RUB,
		TRY: res.Rates.TRY,
	}, nil
}

type usdExchangeRateAPIResponse struct {
	Disclaimer string `json:"disclaimer"`
	License    string `json:"license"`
	Timestamp  int    `json:"timestamp"`
	Base       string `json:"base"`
	Rates      struct {
		RUB float64 `json:"RUB"`
		TRY float64 `json:"TRY"`
	} `json:"rates"`
}

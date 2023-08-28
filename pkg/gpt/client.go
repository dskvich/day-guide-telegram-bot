package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const baseURL = "http://localhost:8080/api/gpt/generate"

type client struct {
	hc *http.Client
}

func NewClient() *client {
	return &client{
		hc: &http.Client{},
	}
}

func (c *client) GetResponse(ctx context.Context, prompt string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parsing base url: %v", err)
	}

	q := u.Query()
	q.Set("prompt", prompt)

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %v", err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %v", err)
	}

	var res response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decoding response body: %v", err)
	}

	return res.Response, nil
}

type response struct {
	Response string `json:"response"`
}

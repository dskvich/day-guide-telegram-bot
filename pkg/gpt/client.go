package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type client struct {
	baseURL *url.URL
	hc      *http.Client
}

func NewClient(base string) (*client, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %v", err)
	}
	return &client{
		baseURL: u,
		hc:      &http.Client{},
	}, nil
}

func (c *client) generateURL() *url.URL {
	const endpointPath = "/api/gpt/generate"
	return c.baseURL.ResolveReference(&url.URL{Path: endpointPath})
}
func (c *client) GetResponse(ctx context.Context, prompt string) (string, error) {
	u := c.generateURL()
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

package googleai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const apiURLChatCompletions = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"

type client struct {
	hc     *http.Client
	apiKey string
}

func NewClient(apiKey string) (*client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}
	return &client{
		apiKey: apiKey,
		hc:     &http.Client{},
	}, nil
}

type geminiRequest struct {
	Contents []domain.GMessage `json:"contents"`
}

type Candidate struct {
	Content struct {
		Parts []domain.GMessagePart `json:"parts"`
	} `json:"content"`
}

type geminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

func (c *client) GenerateResponse(ctx context.Context, messages []domain.GMessage) (domain.GMessage, error) {
	payload := geminiRequest{
		Contents: messages,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return domain.GMessage{}, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	reqURL, err := url.Parse(apiURLChatCompletions)
	if err != nil {
		return domain.GMessage{}, fmt.Errorf("failed to parse API URL: %w", err)
	}
	query := reqURL.Query()
	query.Set("key", c.apiKey)
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return domain.GMessage{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	respBody, err := c.doRequest(req)
	if err != nil {
		return domain.GMessage{}, fmt.Errorf("failed to send chat completion request: %w", err)
	}

	var parsedResp geminiResponse
	if err := json.Unmarshal(respBody, &parsedResp); err != nil {
		return domain.GMessage{}, fmt.Errorf("failed to parse chat completion response: %w", err)
	}

	if len(parsedResp.Candidates) == 0 || len(parsedResp.Candidates[0].Content.Parts) == 0 {
		return domain.GMessage{}, errors.New("no valid response candidates returned")
	}

	return domain.GMessage{
		Role: "model",
		Parts: []domain.GMessagePart{
			{Text: parsedResp.Candidates[0].Content.Parts[0].Text},
		},
	}, nil
}

func (c *client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type client struct {
	token string
	hc    *http.Client
}

func NewClient(token string) (*client, error) {
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}
	return &client{
		token: token,
		hc:    &http.Client{},
	}, nil
}

func (c *client) GenerateTextResponse(systemPrompt, userPrompt string) (string, error) {
	// Prepare the request.
	chatRequest := chatCompletionsRequest{
		Model: "gpt-4-0125-preview",
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: 4096,
	}

	// Send request to the API.
	url := "https://api.openai.com/v1/chat/completions"
	resp, err := c.sendRequest(url, chatRequest)
	if err != nil {
		return "", fmt.Errorf("sending request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Process the response.
	var chatResponse chatCompletionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return "", fmt.Errorf("decoding response data: %v", err)
	}

	if len(chatResponse.Choices) > 0 && fmt.Sprint(chatResponse.Choices[0].Message.Content) != "" {
		return fmt.Sprint(chatResponse.Choices[0].Message.Content), nil
	}

	return "", fmt.Errorf("no completion response from API")
}

func (c *client) sendRequest(url string, chatRequest chatCompletionsRequest) (*http.Response, error) {
	body, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("marshaling chat request: %v", err)
	}

	fmt.Printf("len=%d\n", len(body))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing HTTP request: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

type chatCompletionsRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Content can be a string or a slice
}

type chatCompletionsResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message       chatMessage `json:"message"`
		FinishDetails struct {
			Type string `json:"type"`
			Stop string `json:"stop"`
		} `json:"finish_details"`
		Index int `json:"index"`
	} `json:"choices"`
}

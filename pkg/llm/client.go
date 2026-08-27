package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client interface {
	Generate(ctx context.Context, systemPrompt string) (string, error)
}

type HTTPClient struct {
	URL, APIKey, Model string
	HTTP               *http.Client
}

func (c *HTTPClient) Generate(ctx context.Context, systemPrompt string) (string, error) {
	payload := map[string]any{"model": c.Model, "temperature": 0.7, "messages": []map[string]string{{"role": "system", "content": systemPrompt}, {"role": "user", "content": "Return the question array now."}}}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		response, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm returned status %s: %s", resp.Status, string(response))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}

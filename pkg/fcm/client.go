package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Client struct {
	ProjectID   string
	TokenSource oauth2.TokenSource
	HTTP        *http.Client
}

func NewClient(credentialsFile string, httpClient *http.Client) (*Client, error) {
	credentials, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read Firebase credentials: %w", err)
	}

	config, err := google.JWTConfigFromJSON(credentials, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("parse Firebase credentials: %w", err)
	}

	var metadata struct {
		ProjectID string `json:"project_id"`
	}

	if err := json.Unmarshal(credentials, &metadata); err != nil || metadata.ProjectID == "" {
		return nil, fmt.Errorf("Firebase credentials have no project_id")
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{ProjectID: metadata.ProjectID, TokenSource: config.TokenSource(context.Background()), HTTP: httpClient}, nil
}

func (c *Client) Send(ctx context.Context, token, title, body string) error {
	if c.ProjectID == "" || c.TokenSource == nil {
		return fmt.Errorf("Firebase credentials are not configured")
	}

	payload, err := json.Marshal(map[string]any{"message": map[string]any{
		"token":        token,
		"notification": map[string]string{"title": title, "body": body},
	}})
	if err != nil {
		return err
	}

	accessToken, err := c.TokenSource.Token()
	if err != nil {
		return fmt.Errorf("get Firebase access token: %w", err)
	}

	endpoint := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.ProjectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		response, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("FCM returned status %s: %s", resp.Status, strings.TrimSpace(string(response)))
	}

	return nil
}

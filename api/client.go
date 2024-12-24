package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"outline-cli/config"
)

type Client interface {
	GetDocument(docID string) (*Document, error)
	UpdateDocument(docID string, content string) error
}

// ClientFactory is a function type that creates new API clients
type ClientFactory func(*config.Config) Client

// DefaultClientFactory creates real API clients
var DefaultClientFactory ClientFactory = func(cfg *config.Config) Client {
	return &client{
		httpClient: &http.Client{},
		config:     cfg,
	}
}

// Rename the Client struct to client (private)
type client struct {
	httpClient *http.Client
	config     *config.Config
}

type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Version int    `json:"version"`
}

func (c *client) GetDocument(docID string) (*Document, error) {
	url := fmt.Sprintf("%s/api/documents/%s", c.config.OutlineURL, docID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var doc Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &doc, nil
}

func (c *client) UpdateDocument(docID string, content string) error {
	url := fmt.Sprintf("%s/api/documents/%s", c.config.OutlineURL, docID)

	payload := struct {
		Text string `json:"text"`
	}{
		Text: content,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

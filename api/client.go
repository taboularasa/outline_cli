package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"outline-cli/config"
	"strings"
)

type Client interface {
	GetDocument(docID string, verbose bool) (*Document, error)
	UpdateDocument(docID string, content string, verbose bool) error
	ListDocuments(verbose bool) ([]Document, error)
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

func normalizeURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/")
}

func (c *client) GetDocument(docID string, verbose bool) (*Document, error) {
	url := fmt.Sprintf("%s/api/documents.info", normalizeURL(c.config.OutlineURL))
	if verbose {
		fmt.Printf("Making request to: %s\n", url)
	}

	// Create request body with document ID
	payload := struct {
		ID string `json:"id"`
	}{
		ID: docID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if verbose {
		fmt.Printf("Request headers:\n")
		for k, v := range req.Header {
			fmt.Printf("  %s: %s\n", k, v)
		}
		fmt.Printf("Request body: %s\n", string(body))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if verbose {
		fmt.Printf("Response status: %s\n", resp.Status)
		fmt.Printf("Response body: %s\n", string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		var apiError struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(respBody, &apiError); err == nil {
			return nil, fmt.Errorf("API error: %s - %s", apiError.Error, apiError.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var response struct {
		Data Document `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("decoding response (status %d): %w\nBody: %s", resp.StatusCode, err, string(respBody))
	}

	return &response.Data, nil
}

func (c *client) UpdateDocument(docID string, content string, verbose bool) error {
	url := fmt.Sprintf("%s/api/documents.update", normalizeURL(c.config.OutlineURL))

	payload := struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}{
		ID:   docID,
		Text: content,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if verbose {
		fmt.Printf("Making request to: %s\n", url)
		fmt.Printf("Request payload: %s\n", string(body))
		fmt.Printf("Request headers:\n")
		for k, v := range req.Header {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if verbose {
		fmt.Printf("Response status: %s\n", resp.Status)
		fmt.Printf("Response body: %s\n", string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *client) ListDocuments(verbose bool) ([]Document, error) {
	url := fmt.Sprintf("%s/api/documents.list", normalizeURL(c.config.OutlineURL))
	if verbose {
		fmt.Printf("Making request to: %s\n", url)
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Accept", "application/json")

	if verbose {
		fmt.Printf("Request headers:\n")
		for k, v := range req.Header {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if verbose {
		fmt.Printf("Response status: %s\n", resp.Status)
		fmt.Printf("Response body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []Document `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decoding response (status %d): %w\nBody: %s", resp.StatusCode, err, string(body))
	}

	return response.Data, nil
}

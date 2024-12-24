package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"outline-cli/api"
	"outline-cli/config"
	"strings"

	"github.com/spf13/cobra"
)

var clientFactory api.ClientFactory = api.DefaultClientFactory
var verbose bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "outline",
	Short: "A CLI tool for interacting with Outline API",
	Long: `Outline CLI allows you to manage your Outline documents locally.
You can pull documents, edit them locally, and push changes back to Outline.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

var pullCmd = &cobra.Command{
	Use:   "pull [docID]",
	Short: "Pull a document from Outline",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		client := clientFactory(cfg)
		doc, err := client.GetDocument(args[0], verbose)
		if err != nil {
			return fmt.Errorf("fetching document: %w", err)
		}

		filename := fmt.Sprintf("%s.md", args[0])
		if err := os.WriteFile(filename, []byte(doc.Text), 0644); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}

		fmt.Printf("Successfully pulled document to %s\n", filename)
		return nil
	},
}

var pushCmd = &cobra.Command{
	Use:   "push [docID]",
	Short: "Push local changes to Outline",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		client := clientFactory(cfg)
		filename := fmt.Sprintf("%s.md", args[0])
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		if err := client.UpdateDocument(args[0], string(content), verbose); err != nil {
			return fmt.Errorf("updating document: %w", err)
		}

		fmt.Printf("Successfully pushed changes to document %s\n", args[0])
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [docID]",
	Short: "Compare local and remote versions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement diff logic
		return nil
	},
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Print debug information",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		fmt.Printf("Configuration:\n")
		fmt.Printf("  Outline URL: %s\n", cfg.OutlineURL)
		fmt.Printf("  API Key: %s\n", maskAPIKey(cfg.APIKey))
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available documents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		client := clientFactory(cfg)
		docs, err := client.ListDocuments(verbose)
		if err != nil {
			return fmt.Errorf("listing documents: %w", err)
		}

		for _, doc := range docs {
			fmt.Printf("%s: %s\n", doc.ID, doc.Title)
		}
		return nil
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test API connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		url := fmt.Sprintf("%s/api/auth.info", normalizeURL(cfg.OutlineURL))

		// Create an empty payload since it's a POST request
		payload := struct{}{}
		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshaling payload: %w", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		if verbose {
			fmt.Printf("Making request to: %s\n", url)
			fmt.Printf("Request headers:\n")
			for k, v := range req.Header {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if verbose {
			fmt.Printf("Response status: %s\n", resp.Status)
			fmt.Printf("Response body: %s\n", string(body))
		}

		if resp.StatusCode != http.StatusOK {
			var apiError struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(body, &apiError); err == nil {
				return fmt.Errorf("API error: %s - %s", apiError.Error, apiError.Message)
			}
			return fmt.Errorf("API error: %s", string(body))
		}

		fmt.Println("API connection successful!")
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [docID]",
	Short: "Update document metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		url := fmt.Sprintf("%s/api/documents.update", normalizeURL(cfg.OutlineURL))
		payload := struct {
			ID      string `json:"id"`
			Publish bool   `json:"publish"`
		}{
			ID:      args[0],
			Publish: true,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshaling payload: %w", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		fmt.Printf("Successfully updated document %s\n", args[0])
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		client := clientFactory(cfg)
		doc, err := client.CreateDocument(
			args[0],
			"# "+args[0]+"\n\nNew document created via CLI.",
			"8f2de8e6-a423-4960-8802-18c0da301989", // Infrastructure collection ID
			verbose,
		)
		if err != nil {
			return fmt.Errorf("creating document: %w", err)
		}

		fmt.Printf("Successfully created document with ID: %s\n", doc.ID)
		return nil
	},
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func normalizeURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/")
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(diffCmd)
	RootCmd.AddCommand(debugCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(testCmd)
	RootCmd.AddCommand(updateCmd)
	RootCmd.AddCommand(createCmd)
}

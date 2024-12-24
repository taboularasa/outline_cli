package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"outline-cli/api"
	"outline-cli/config"

	"github.com/spf13/cobra"
)

// Helper function to silence cobra command output during tests
func silenceOutput(_ *testing.T) func() {
	null, _ := os.Open(os.DevNull)
	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout = null
	os.Stderr = null
	return func() {
		os.Stdout = stdout
		os.Stderr = stderr
		null.Close()
	}
}

// Mock config for testing
var testConfig = &config.Config{
	APIKey:     "test-key",
	OutlineURL: "https://test.outline.com",
}

// Mock the config loading
func mockConfigLoader(cfg *config.Config) func() {
	original := config.LoadConfig
	config.LoadConfig = func() (*config.Config, error) {
		return cfg, nil
	}
	return func() {
		config.LoadConfig = original
	}
}

func TestPullCommand(t *testing.T) {
	// Silence cobra output
	cleanup := silenceOutput(t)
	defer cleanup()

	// Mock config loading at the start
	cleanupConfig := mockConfigLoader(testConfig)
	defer cleanupConfig()

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "outline-cli-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temporary directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name        string
		docID       string
		mockClient  api.Client
		wantErr     bool
		wantErrMsg  string
		wantContent string
	}{
		{
			name:  "successful pull",
			docID: "doc123",
			mockClient: &api.MockClient{
				GetDocumentFunc: func(docID string, verbose bool) (*api.Document, error) {
					return &api.Document{
						ID:    docID,
						Title: "Test Doc",
						Text:  "# Test Content",
					}, nil
				},
			},
			wantContent: "# Test Content",
		},
		{
			name:  "api error",
			docID: "doc456",
			mockClient: &api.MockClient{
				GetDocumentFunc: func(docID string, verbose bool) (*api.Document, error) {
					return nil, fmt.Errorf("API error")
				},
			},
			wantErr:    true,
			wantErrMsg: "fetching document: API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := &cobra.Command{
				Use:   pullCmd.Use,
				Short: pullCmd.Short,
				Args:  pullCmd.Args,
				RunE:  pullCmd.RunE,
			}
			cmd.SetArgs([]string{tt.docID})

			// Override the client factory
			originalFactory := clientFactory
			clientFactory = func(cfg *config.Config) api.Client {
				return tt.mockClient
			}
			defer func() { clientFactory = originalFactory }()

			// Execute the command
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("got error message = %q, want %q", err.Error(), tt.wantErrMsg)
				return
			}

			// Check file content only for successful cases
			if !tt.wantErr {
				filename := fmt.Sprintf("%s.md", tt.docID)
				content, err := os.ReadFile(filename)
				if err != nil {
					t.Fatalf("failed to read output file: %v", err)
				}
				if string(content) != tt.wantContent {
					t.Errorf("got content = %q, want %q", string(content), tt.wantContent)
				}
			}
		})
	}
}

func TestPushCommand(t *testing.T) {
	// Silence cobra output
	cleanup := silenceOutput(t)
	defer cleanup()

	// Mock config loading at the start
	cleanupConfig := mockConfigLoader(testConfig)
	defer cleanupConfig()

	tmpDir, err := os.MkdirTemp("", "outline-cli-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name         string
		docID        string
		fileExists   bool
		content      string
		mockClient   api.Client
		wantErr      bool
		wantErrMsg   string
		errMsgPrefix bool
	}{
		{
			name:       "successful push",
			docID:      "doc123",
			fileExists: true,
			content:    "# Updated Content",
			mockClient: &api.MockClient{
				UpdateDocumentFunc: func(docID string, content string, verbose bool) error {
					if content != "# Updated Content" {
						return fmt.Errorf("unexpected content: %s", content)
					}
					return nil
				},
			},
		},
		{
			name:       "missing file",
			docID:      "doc456",
			fileExists: false,
			mockClient: &api.MockClient{
				UpdateDocumentFunc: func(docID string, content string, verbose bool) error {
					return nil
				},
			},
			wantErr:      true,
			wantErrMsg:   "reading file: ",
			errMsgPrefix: true,
		},
		{
			name:       "api error",
			docID:      "doc789",
			fileExists: true,
			content:    "# Content",
			mockClient: &api.MockClient{
				UpdateDocumentFunc: func(docID string, content string, verbose bool) error {
					return fmt.Errorf("API error")
				},
			},
			wantErr:    true,
			wantErrMsg: "updating document: API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file if needed
			if tt.fileExists {
				filename := fmt.Sprintf("%s.md", tt.docID)
				if err := os.WriteFile(filename, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			// Create a fresh command for each test
			cmd := &cobra.Command{
				Use:   pushCmd.Use,
				Short: pushCmd.Short,
				Args:  pushCmd.Args,
				RunE:  pushCmd.RunE,
			}
			cmd.SetArgs([]string{tt.docID})

			// Override the client factory
			originalFactory := clientFactory
			clientFactory = func(cfg *config.Config) api.Client {
				return tt.mockClient
			}
			defer func() { clientFactory = originalFactory }()

			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				gotMsg := err.Error()
				if tt.errMsgPrefix {
					if !strings.HasPrefix(gotMsg, tt.wantErrMsg) {
						t.Errorf("error message = %q, want prefix %q", gotMsg, tt.wantErrMsg)
					}
				} else if gotMsg != tt.wantErrMsg {
					t.Errorf("error message = %q, want %q", gotMsg, tt.wantErrMsg)
				}
			}
		})
	}
}

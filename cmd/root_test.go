package cmd

import (
	"fmt"
	"os"
	"testing"

	"outline-cli/api"
	"outline-cli/config"
)

// Helper function to silence command output during tests
func silenceOutput(t *testing.T) func() {
	t.Helper()
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = null
	os.Stderr = null
	return func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		null.Close()
	}
}

// Mock config for testing
var testConfig = &config.Config{
	APIKey:     "test-key",
	OutlineURL: "https://test.outline.com",
}

// Mock the config loading
func mockConfigLoader() func() {
	original := config.LoadConfig
	config.LoadConfig = func() (*config.Config, error) {
		return testConfig, nil
	}
	return func() {
		config.LoadConfig = original
	}
}

// Reset commands before each test
func resetCommands() {
	// Reset the commands
	pullCmd.ResetFlags()
	pushCmd.ResetFlags()
	createCmd.ResetFlags()

	// Reset the root command and its flags
	RootCmd.ResetFlags()
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	// Re-add commands to root
	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(createCmd)
}

type mockClient struct {
	documents map[string]*api.Document
}

func newMockClient() *mockClient {
	return &mockClient{
		documents: make(map[string]*api.Document),
	}
}

func (m *mockClient) GetDocument(docID string, verbose bool) (*api.Document, error) {
	doc, exists := m.documents[docID]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}
	return doc, nil
}

func (m *mockClient) UpdateDocument(docID string, content string, verbose bool) error {
	doc, exists := m.documents[docID]
	if !exists {
		return fmt.Errorf("document not found")
	}
	doc.Text = content
	return nil
}

func (m *mockClient) ListDocuments(verbose bool) ([]api.Document, error) {
	docs := make([]api.Document, 0, len(m.documents))
	for _, doc := range m.documents {
		docs = append(docs, *doc)
	}
	return docs, nil
}

func (m *mockClient) CreateDocument(title string, text string, collectionId string, verbose bool) (*api.Document, error) {
	doc := &api.Document{
		ID:      "test-doc-id",
		Title:   title,
		Text:    text,
		Version: 1,
	}
	m.documents["test-doc-id"] = doc
	return doc, nil
}

func TestPullCommand(t *testing.T) {
	cleanup := silenceOutput(t)
	defer cleanup()

	// Reset commands
	resetCommands()

	// Setup mock config and client
	configCleanup := mockConfigLoader()
	defer configCleanup()

	mock := newMockClient()
	mock.documents["test-id"] = &api.Document{
		ID:    "test-id",
		Title: "Test Document",
		Text:  "Test content",
	}

	clientFactory = func(_ *config.Config) api.Client {
		return mock
	}

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temporary directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Execute pull command
	RootCmd.SetArgs([]string{"pull", "test-id"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created with correct content
	content, err := os.ReadFile("test-id.md")
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	if string(content) != "Test content" {
		t.Errorf("expected content %q, got %q", "Test content", string(content))
	}
}

func TestPushCommand(t *testing.T) {
	cleanup := silenceOutput(t)
	defer cleanup()

	// Reset commands
	resetCommands()

	// Setup mock config and client
	configCleanup := mockConfigLoader()
	defer configCleanup()

	mock := newMockClient()
	mock.documents["test-id"] = &api.Document{
		ID:    "test-id",
		Title: "Test Document",
		Text:  "Original content",
	}

	clientFactory = func(_ *config.Config) api.Client {
		return mock
	}

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temporary directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create test file
	if err := os.WriteFile("test-id.md", []byte("Updated content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Execute push command
	RootCmd.SetArgs([]string{"push", "test-id"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify document was updated
	doc := mock.documents["test-id"]
	if doc == nil {
		t.Fatal("document was not created")
	}
	if doc.Text != "Updated content" {
		t.Errorf("expected content %q, got %q", "Updated content", doc.Text)
	}
}

func TestCreateCommand(t *testing.T) {
	cleanup := silenceOutput(t)
	defer cleanup()

	// Reset commands
	resetCommands()

	// Setup mock config and client
	configCleanup := mockConfigLoader()
	defer configCleanup()

	mock := newMockClient()
	clientFactory = func(_ *config.Config) api.Client {
		return mock
	}

	// Execute create command
	RootCmd.SetArgs([]string{"create", "New Test Document"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify document was created
	doc := mock.documents["test-doc-id"]
	if doc == nil {
		t.Fatal("document was not created")
	}
	if doc.Title != "New Test Document" {
		t.Errorf("expected title %q, got %q", "New Test Document", doc.Title)
	}
}

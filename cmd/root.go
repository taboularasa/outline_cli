package cmd

import (
	"fmt"
	"os"
	"outline-cli/api"
	"outline-cli/config"

	"github.com/spf13/cobra"
)

var clientFactory api.ClientFactory = api.DefaultClientFactory

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
		doc, err := client.GetDocument(args[0])
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

		if err := client.UpdateDocument(args[0], string(content)); err != nil {
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

func init() {
	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(diffCmd)
}

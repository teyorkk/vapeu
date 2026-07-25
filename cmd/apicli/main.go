package main

import (
	"fmt"
	"os"

	"vapeu/internal/api"
	"vapeu/internal/impexp"
	"vapeu/internal/models"
	"vapeu/internal/storage"
	"vapeu/internal/theme"
	"vapeu/internal/ui"
	"vapeu/internal/variables"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	configDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "apicli",
		Short: "A cross-platform keyboard-driven terminal API client",
		Long:  `apicli is a lightweight terminal-based API client for testing and debugging HTTP APIs locally or over SSH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := storage.NewStorage(configDir)
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}

			app := ui.NewAppModel(st)
			p := tea.NewProgram(app, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running TUI: %w", err)
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configDir, "config-dir", "c", "", "Path to storage directory (default ~/.apiclient)")

	// Subcommand: run
	runCmd := &cobra.Command{
		Use:   "run <url-or-curl>",
		Short: "Execute an HTTP request directly from the terminal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := storage.NewStorage(configDir)
			if err != nil {
				return err
			}

			cfg, _ := st.LoadConfig()
			target := args[0]

			var req *models.Request
			if len(target) > 4 && target[:4] == "curl" {
				parsedReq, err := impexp.ParseCurl(target)
				if err != nil {
					return err
				}
				req = parsedReq
			} else {
				req = &models.Request{
					Method: "GET",
					URL:    target,
					Auth:   models.AuthConfig{Type: models.AuthNone},
					Body:   models.RequestBody{Type: models.BodyNone},
				}
			}

			client := api.NewClient(api.ClientOptions{
				TimeoutSec:  cfg.DefaultTimeoutSec,
				InsecureSSL: !cfg.SSLVerification,
			})

			resolver := variables.NewResolver(nil, nil, nil, nil)
			resp, err := client.Execute(req, resolver)
			if err != nil {
				return err
			}

			if resp.Error != "" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
				os.Exit(1)
			}

			if pretty, err := theme.PrettyFormatJSON(resp.Body); err == nil {
				fmt.Println(pretty)
			} else {
				fmt.Println(string(resp.Body))
			}
			return nil
		},
	}

	// Subcommand: import
	importCmd := &cobra.Command{
		Use:   "import <filepath>",
		Short: "Import OpenAPI or Postman collection into apicli",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := storage.NewStorage(configDir)
			if err != nil {
				return err
			}

			filePath := args[0]
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			col, err := impexp.ImportOpenAPI(data)
			if err != nil || len(col.Nodes) == 0 {
				col, err = impexp.ImportPostmanCollection(data)
				if err != nil {
					return fmt.Errorf("failed to import collection: %w", err)
				}
			}

			if err := st.SaveCollection(*col); err != nil {
				return fmt.Errorf("failed to save collection: %w", err)
			}

			fmt.Printf("Successfully imported collection '%s' with %d endpoints.\n", col.Name, len(col.Nodes))
			return nil
		},
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(importCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

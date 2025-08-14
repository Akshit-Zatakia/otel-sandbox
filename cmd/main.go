package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Akshit-Zatakia/otel-sandbox/internal/export"
	"github.com/Akshit-Zatakia/otel-sandbox/internal/manager"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "otel-sandbox",
		Short: "Local OpenTelemetry sandbox (lightweight, no Docker)",
	}

	rootCmd.AddCommand(upCmd())
	rootCmd.AddCommand(downCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(downloadCmd())
	rootCmd.AddCommand(verifyCmd())
	rootCmd.AddCommand(exportCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Add the export command function
func exportCmd() *cobra.Command {
	var format, output, service string
	var fromTime, toTime string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export collected telemetry data",
		Long: `Export telemetry data collected by the OTel Collector.

Supported formats:
  - json: Full JSON export (default)
  - summary: Human-readable summary
  - csv: Comma-separated values

Examples:
  otel-sandbox export
  otel-sandbox export --format summary
  otel-sandbox export --format json --output data.json
  otel-sandbox export --service my-service --format summary`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := export.ExportConfig{
				OutputFormat: format,
				OutputFile:   output,
				ServiceName:  service,
			}

			// Parse time filters if provided
			if fromTime != "" {
				if t, err := time.Parse(time.RFC3339, fromTime); err == nil {
					config.TimeFrom = &t
				} else {
					return fmt.Errorf("invalid from time format. Use RFC3339 format like: 2023-01-01T00:00:00Z")
				}
			}
			if toTime != "" {
				if t, err := time.Parse(time.RFC3339, toTime); err == nil {
					config.TimeTo = &t
				} else {
					return fmt.Errorf("invalid to time format. Use RFC3339 format like: 2023-01-01T23:59:59Z")
				}
			}

			return export.Export(config)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format (json, csv, summary)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&service, "service", "", "Filter by service name")
	cmd.Flags().StringVar(&fromTime, "from", "", "Filter from time (RFC3339 format)")
	cmd.Flags().StringVar(&toTime, "to", "", "Filter to time (RFC3339 format)")

	return cmd
}

func upCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start Collector + Jaeger + Prometheus",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := manager.NewManager("./state.json")
			return mgr.Up()
		},
	}
	return cmd
}

func downCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Stop all sandbox processes",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := manager.NewManager("./state.json")
			return mgr.Down()
		},
	}
	return cmd
}

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of sandbox processes",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := manager.NewManager("./state.json")
			return mgr.Status()
		},
	}
	return cmd
}

func downloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download binaries for Collector + Jaeger + Prometheus",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := manager.NewManager("./state.json")
			return mgr.DownloadBinaries()
		},
	}
	return cmd
}

func verifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify Collector + Jaeger + Prometheus",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := manager.NewManager("./state.json")
			return mgr.Verify()
		},
	}
	return cmd
}

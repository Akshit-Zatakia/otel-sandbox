package main

import (
	"fmt"
	"os"

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

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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

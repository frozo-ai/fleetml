package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fleetml",
	Short: "FleetML - Edge MLOps Platform",
	Long:  "FleetML deploys, updates, and monitors ML models across heterogeneous edge device fleets.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "FleetML server address")
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication")
	rootCmd.PersistentFlags().String("format", "table", "Output format (table, json, yaml)")
}

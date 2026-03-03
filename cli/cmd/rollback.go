package cmd

import (
	"fmt"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback <deployment_id>",
	Short: "Rollback a deployment to the previous model version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]

		serverURL, _ := cmd.Flags().GetString("server")
		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}

		apiClient := client.NewAPIClient(serverURL, "")
		if err := apiClient.Login("admin@fleetml.io", "admin123"); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		result, err := apiClient.RollbackDeployment(deploymentID)
		if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		fmt.Printf("Rollback deployment created: %v\n", result["id"])
		fmt.Printf("State: %v\n", result["state"])
		fmt.Printf("Devices: %v\n", result["total_devices"])

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

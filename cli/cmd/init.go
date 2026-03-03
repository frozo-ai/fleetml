package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize FleetML configuration",
	Long:  "Creates a config file and verifies server connectivity.",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")

		// Test connectivity
		c := client.NewAPIClient(serverURL, apiKey)
		health, err := c.HealthCheck()
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}

		fmt.Printf("Connected to FleetML server %s\n", health["version"])

		// Save config
		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, ".fleetml")
		os.MkdirAll(configDir, 0o755)

		config := map[string]interface{}{
			"server": map[string]string{
				"address": serverURL,
				"api_key": apiKey,
			},
		}

		configPath := filepath.Join(configDir, "config.yaml")
		data, _ := yaml.Marshal(config)
		if err := os.WriteFile(configPath, data, 0o600); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		fmt.Printf("Config saved to %s\n", configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

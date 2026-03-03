package cmd

import (
	"fmt"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/fleetml/fleetml/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show fleet status",
	Long:  "Display the current status of devices in the fleet.",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		fleetName, _ := cmd.Flags().GetString("fleet")
		format, _ := cmd.Flags().GetString("format")

		c := client.NewAPIClient(serverURL, apiKey)
		result, err := c.ListDevices(fleetName)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		if format == "json" {
			output.PrintJSON(result)
			return nil
		}

		// Table format
		devices, ok := result["devices"].([]interface{})
		if !ok || len(devices) == 0 {
			fmt.Println("No devices found.")
			return nil
		}

		headers := []string{"DEVICE", "STATUS", "ARCH", "GPU", "RUNTIME", "RAM (MB)"}
		var rows [][]string

		for _, d := range devices {
			dev, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			rows = append(rows, []string{
				fmt.Sprintf("%v", dev["device_id"]),
				fmt.Sprintf("%v", dev["status"]),
				fmt.Sprintf("%v", dev["arch"]),
				fmt.Sprintf("%v", dev["gpu_type"]),
				fmt.Sprintf("%v", dev["runtime"]),
				fmt.Sprintf("%v", dev["ram_mb"]),
			})
		}

		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	statusCmd.Flags().String("fleet", "", "Filter by fleet")
	statusCmd.Flags().String("device", "", "Show single device detail")

	rootCmd.AddCommand(statusCmd)
}

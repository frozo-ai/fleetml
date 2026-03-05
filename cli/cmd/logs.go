package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <device_id>",
	Short: "Stream device logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceID := args[0]
		follow, _ := cmd.Flags().GetBool("follow")
		since, _ := cmd.Flags().GetString("since")
		level, _ := cmd.Flags().GetString("level")
		limit, _ := cmd.Flags().GetInt("limit")

		serverURL, _ := cmd.Root().PersistentFlags().GetString("server")
		apiKey, _ := cmd.Root().PersistentFlags().GetString("api-key")
		apiClient := client.NewAPIClient(serverURL, apiKey)

		if follow {
			// Streaming mode
			fmt.Fprintf(os.Stderr, "Streaming logs for device %s", deviceID)
			if level != "" {
				fmt.Fprintf(os.Stderr, " (level: %s)", level)
			}
			fmt.Fprintln(os.Stderr)

			stream, err := apiClient.StreamDeviceLogs(deviceID, since, level)
			if err != nil {
				return fmt.Errorf("failed to stream logs: %w", err)
			}
			defer stream.Close()

			scanner := bufio.NewScanner(stream)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
			return scanner.Err()
		}

		// Batch mode — fetch recent logs
		fmt.Fprintf(os.Stderr, "Fetching logs for device %s", deviceID)
		if since != "" {
			fmt.Fprintf(os.Stderr, " (since %s)", since)
		}
		if level != "" {
			fmt.Fprintf(os.Stderr, " (level: %s)", level)
		}
		fmt.Fprintln(os.Stderr)

		logs, err := apiClient.GetDeviceLogs(deviceID, since, level, limit)
		if err != nil {
			return fmt.Errorf("failed to fetch logs: %w", err)
		}

		if len(logs) == 0 {
			fmt.Println("No logs found.")
			return nil
		}

		for _, entry := range logs {
			ts, _ := entry["timestamp"].(string)
			lvl, _ := entry["level"].(string)
			msg, _ := entry["message"].(string)
			fmt.Printf("%s [%s] %s\n", ts, lvl, msg)
		}

		fmt.Fprintf(os.Stderr, "\n%d log entries\n", len(logs))
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Stream logs in real time")
	logsCmd.Flags().String("since", "", "Show logs since (e.g., 1h, 24h, 7d)")
	logsCmd.Flags().String("level", "", "Filter by level (debug, info, warn, error)")
	logsCmd.Flags().Int("limit", 100, "Max lines")

	rootCmd.AddCommand(logsCmd)
}

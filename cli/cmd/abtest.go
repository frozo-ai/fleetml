package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/fleetml/fleetml/cli/internal/output"
	"github.com/spf13/cobra"
)

var abtestCmd = &cobra.Command{
	Use:   "ab-test",
	Short: "Manage A/B tests",
	Long:  "Create, monitor, and stop A/B tests between model variants.",
}

var abtestCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new A/B test",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")

		name, _ := cmd.Flags().GetString("name")
		modelA, _ := cmd.Flags().GetString("model-a")
		modelB, _ := cmd.Flags().GetString("model-b")
		splitStr, _ := cmd.Flags().GetString("split")
		fleet, _ := cmd.Flags().GetString("fleet")
		metric, _ := cmd.Flags().GetString("metric")
		duration, _ := cmd.Flags().GetString("duration")
		autoPromote, _ := cmd.Flags().GetBool("auto-promote")

		if name == "" || modelA == "" || modelB == "" {
			return fmt.Errorf("--name, --model-a, and --model-b are required")
		}

		// Parse split (e.g., "80/20")
		splitA, splitB := 80, 20
		if splitStr != "" {
			parts := strings.Split(splitStr, "/")
			if len(parts) == 2 {
				fmt.Sscanf(parts[0], "%d", &splitA)
				fmt.Sscanf(parts[1], "%d", &splitB)
			}
		}

		c := client.NewAPIClient(serverURL, apiKey)

		result, err := c.CreateABTest(name, modelA, modelB, splitA, splitB, fleet, metric, duration, autoPromote)
		if err != nil {
			return fmt.Errorf("failed to create A/B test: %w", err)
		}

		output.Success(fmt.Sprintf("A/B test created: %v", result["id"]))
		fmt.Printf("  Name:    %v\n", result["name"])
		fmt.Printf("  Split:   %d/%d (A/B)\n", splitA, splitB)
		fmt.Printf("  Metric:  %v\n", result["metric"])
		fmt.Printf("  State:   %v\n", result["state"])

		return nil
	},
}

var abtestListCmd = &cobra.Command{
	Use:   "list",
	Short: "List A/B tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		state, _ := cmd.Flags().GetString("state")

		c := client.NewAPIClient(serverURL, apiKey)

		result, err := c.ListABTests(state)
		if err != nil {
			return fmt.Errorf("failed to list A/B tests: %w", err)
		}

		tests, ok := result["ab_tests"].([]interface{})
		if !ok || len(tests) == 0 {
			fmt.Println("No A/B tests found.")
			return nil
		}

		fmt.Printf("%-36s  %-20s  %-10s  %-8s  %-10s\n", "ID", "NAME", "SPLIT", "METRIC", "STATE")
		fmt.Println(strings.Repeat("-", 90))

		for _, t := range tests {
			test, _ := t.(map[string]interface{})
			splitA := toInt(test["split_a"])
			splitB := toInt(test["split_b"])
			fmt.Printf("%-36v  %-20v  %d/%-7d  %-8v  %-10v\n",
				test["id"], test["name"], splitA, splitB, test["metric"], test["state"])
		}

		return nil
	},
}

var abtestStatusCmd = &cobra.Command{
	Use:   "status <test-id>",
	Short: "Show A/B test details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")

		c := client.NewAPIClient(serverURL, apiKey)

		result, err := c.GetABTest(args[0])
		if err != nil {
			return fmt.Errorf("failed to get A/B test: %w", err)
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var abtestStopCmd = &cobra.Command{
	Use:   "stop <test-id>",
	Short: "Stop a running A/B test",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		winner, _ := cmd.Flags().GetString("winner")

		c := client.NewAPIClient(serverURL, apiKey)

		result, err := c.StopABTest(args[0], winner)
		if err != nil {
			return fmt.Errorf("failed to stop A/B test: %w", err)
		}

		output.Success(fmt.Sprintf("A/B test stopped: %v", result["id"]))
		if w, ok := result["winner"]; ok && w != nil {
			fmt.Printf("  Winner: Model %v\n", w)
		}
		return nil
	},
}

func init() {
	// Create flags
	abtestCreateCmd.Flags().String("name", "", "Test name (required)")
	abtestCreateCmd.Flags().String("model-a", "", "Model A ID (required)")
	abtestCreateCmd.Flags().String("model-b", "", "Model B ID (required)")
	abtestCreateCmd.Flags().String("split", "80/20", "Traffic split (A/B)")
	abtestCreateCmd.Flags().String("fleet", "", "Target fleet ID")
	abtestCreateCmd.Flags().String("metric", "accuracy", "Comparison metric")
	abtestCreateCmd.Flags().String("duration", "", "Test duration (e.g., 1h, 24h)")
	abtestCreateCmd.Flags().Bool("auto-promote", false, "Auto-promote winner")

	// List flags
	abtestListCmd.Flags().String("state", "", "Filter by state (running, stopped, completed)")

	// Stop flags
	abtestStopCmd.Flags().String("winner", "", "Declare winner (a or b)")

	abtestCmd.AddCommand(abtestCreateCmd)
	abtestCmd.AddCommand(abtestListCmd)
	abtestCmd.AddCommand(abtestStatusCmd)
	abtestCmd.AddCommand(abtestStopCmd)

	rootCmd.AddCommand(abtestCmd)
}

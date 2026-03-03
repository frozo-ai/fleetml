package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/fleetml/fleetml/cli/internal/output"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <model_file>",
	Short: "Deploy a model to devices",
	Long:  "Upload a model file and deploy it to a fleet of devices.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelFile := args[0]
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")

		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		fleetName, _ := cmd.Flags().GetString("fleet")
		device, _ := cmd.Flags().GetString("device")
		policy, _ := cmd.Flags().GetString("policy")
		wait, _ := cmd.Flags().GetBool("wait")

		// Infer name from filename if not provided
		if name == "" {
			base := filepath.Base(modelFile)
			name = strings.TrimSuffix(base, filepath.Ext(base))
		}
		if version == "" {
			version = "1.0"
		}

		// Detect format from extension
		format := strings.TrimPrefix(filepath.Ext(modelFile), ".")
		if format == "" {
			format = "onnx"
		}

		c := client.NewAPIClient(serverURL, apiKey)

		// 1. Upload model
		fmt.Printf("Uploading %s...\n", modelFile)
		m, err := c.UploadModel(modelFile, name, version, format)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		output.Success(fmt.Sprintf("Registered as %s v%s", name, version))

		// 2. Deploy
		targetType := "fleet"
		targetID := fleetName
		if device != "" {
			targetType = "device"
			targetID = device
		}

		d, err := c.CreateDeployment(name, version, targetType, targetID, policy)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		deployID := fmt.Sprintf("%v", d["id"])
		totalDevices := d["total_devices"]
		output.Success(fmt.Sprintf("Rolling out to %v devices (policy: %s)", totalDevices, policy))

		// 3. Wait with progress bar if requested
		if wait {
			timeout, _ := cmd.Flags().GetDuration("timeout")
			if err := waitForDeployment(c, deployID, timeout); err != nil {
				return err
			}
		}

		fmt.Printf("Deployment ID: %s\n", deployID)
		_ = m
		return nil
	},
}

func waitForDeployment(c *client.APIClient, deployID string, timeout time.Duration) error {
	fmt.Println("Waiting for deployment to complete...")
	deadline := time.Now().Add(timeout)
	lastPrinted := ""

	for time.Now().Before(deadline) {
		status, err := c.GetDeployment(deployID)
		if err != nil {
			return fmt.Errorf("failed to get deployment status: %w", err)
		}

		state := fmt.Sprintf("%v", status["state"])
		total := toInt(status["total_devices"])
		completed := toInt(status["completed_devices"])
		failed := toInt(status["failed_devices"])
		queued := toInt(status["queued_devices"])

		// Build progress bar
		progressLine := renderProgress(state, total, completed, failed, queued)
		if progressLine != lastPrinted {
			fmt.Printf("\r\033[K%s", progressLine)
			lastPrinted = progressLine
		}

		switch state {
		case "completed":
			fmt.Printf("\n")
			output.Success(fmt.Sprintf("Deployment completed: %d/%d devices", completed, total))
			return nil
		case "failed":
			fmt.Printf("\n")
			errMsg := ""
			if e, ok := status["error"]; ok {
				errMsg = fmt.Sprintf(": %v", e)
			}
			return fmt.Errorf("deployment failed (%d/%d succeeded, %d failed)%s",
				completed, total, failed, errMsg)
		case "cancelled":
			fmt.Printf("\n")
			return fmt.Errorf("deployment was cancelled")
		case "rolled_back":
			fmt.Printf("\n")
			return fmt.Errorf("deployment was rolled back (canary check failed)")
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Printf("\n")
	return fmt.Errorf("timeout waiting for deployment after %v", timeout)
}

func renderProgress(state string, total, completed, failed, queued int) string {
	if total == 0 {
		return fmt.Sprintf("[%s] waiting...", state)
	}

	active := total - completed - failed - queued
	pct := float64(completed) / float64(total) * 100

	// Build bar
	barWidth := 30
	filledWidth := int(float64(barWidth) * float64(completed) / float64(total))
	failedWidth := int(float64(barWidth) * float64(failed) / float64(total))
	emptyWidth := barWidth - filledWidth - failedWidth

	if emptyWidth < 0 {
		emptyWidth = 0
	}

	bar := strings.Repeat("=", filledWidth) +
		strings.Repeat("!", failedWidth) +
		strings.Repeat("-", emptyWidth)

	return fmt.Sprintf("[%s] [%s] %.0f%% (%d/%d done, %d failed, %d active, %d queued)",
		state, bar, pct, completed, total, failed, active, queued)
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	default:
		return 0
	}
}

func init() {
	deployCmd.Flags().String("name", "", "Model name")
	deployCmd.Flags().String("version", "", "Model version")
	deployCmd.Flags().String("fleet", "default", "Target fleet name")
	deployCmd.Flags().String("device", "", "Target single device ID")
	deployCmd.Flags().String("labels", "", "Label selector (key=value)")
	deployCmd.Flags().String("policy", "immediate", "Deployment policy (immediate, canary)")
	deployCmd.Flags().String("canary-stages", "", "Canary config (e.g. '5:5m,50:10m,100:15m')")
	deployCmd.Flags().Bool("wait", false, "Wait for deployment to complete")
	deployCmd.Flags().Duration("timeout", 10*time.Minute, "Wait timeout")

	rootCmd.AddCommand(deployCmd)
}

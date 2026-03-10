package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// cloudServerURL is the FleetML Cloud SaaS endpoint.
// Override with FLEETML_CLOUD_URL env var for custom deployments.
var cloudServerURL = getCloudURL()

func getCloudURL() string {
	if v := os.Getenv("FLEETML_CLOUD_URL"); v != "" {
		return v
	}
	return "https://server-production-91d4.up.railway.app"
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize FleetML configuration",
	Long:  "Creates a config file and verifies server connectivity.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cloud, _ := cmd.Flags().GetBool("cloud")

		if cloud {
			return initCloud(cmd)
		}

		return initSelfHosted(cmd)
	},
}

func init() {
	initCmd.Flags().Bool("cloud", false, "Connect to FleetML Cloud (hosted SaaS)")
	rootCmd.AddCommand(initCmd)
}

func initSelfHosted(cmd *cobra.Command) error {
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
	return saveConfig(serverURL, apiKey, "")
}

func initCloud(cmd *cobra.Command) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("FleetML Cloud Setup")
	fmt.Println("====================")
	fmt.Println()

	// Check if already configured
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".fleetml", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Existing config found at %s\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Try login first
	token, err := cloudLogin(email, password)
	if err != nil {
		// Try registration
		fmt.Println("Account not found. Creating a new account...")
		fmt.Print("Your name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("Organization name (optional): ")
		org, _ := reader.ReadString('\n')
		org = strings.TrimSpace(org)

		if err := cloudRegister(email, password, name, org); err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}

		token, err = cloudLogin(email, password)
		if err != nil {
			return fmt.Errorf("login after registration failed: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("Authenticated successfully!")

	// Save config
	if err := saveConfig(cloudServerURL, token, "cloud"); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("FleetML Cloud is ready! Next steps:")
	fmt.Println()
	fmt.Println("  1. Install the agent on your edge device:")
	fmt.Println("     curl -sSL https://raw.githubusercontent.com/ashish-frozo/fleetML/main/scripts/install-agent.sh | sh")
	fmt.Println()
	fmt.Println("  2. Deploy a model:")
	fmt.Println("     fleetml deploy model.onnx --fleet default")
	fmt.Println()
	fmt.Println("  3. Check fleet status:")
	fmt.Println("     fleetml status")
	fmt.Println()

	return nil
}

func cloudLogin(email, password string) (string, error) {
	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	resp, err := http.Post(cloudServerURL+"/api/v1/auth/login", "application/json", strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed (status %d)", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.Token, nil
}

func cloudRegister(email, password, name, org string) error {
	payload := map[string]string{
		"email":        email,
		"password":     password,
		"name":         name,
		"organization": org,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(cloudServerURL+"/api/v1/auth/register", "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("%s", errResp.Error)
	}

	return nil
}

func saveConfig(serverURL, token, mode string) error {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".fleetml")
	os.MkdirAll(configDir, 0o755)

	config := map[string]interface{}{
		"server": map[string]string{
			"address": serverURL,
			"api_key": token,
		},
	}
	if mode != "" {
		config["mode"] = mode
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, _ := yaml.Marshal(config)
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Config saved to %s\n", configPath)
	return nil
}

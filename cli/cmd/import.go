package cmd

import (
	"fmt"
	"strings"

	"github.com/fleetml/fleetml/cli/internal/client"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import models from external registries",
}

var importMLflowCmd = &cobra.Command{
	Use:   "mlflow",
	Short: "Import a model from MLflow registry",
	Example: `  fleetml import mlflow --model my-model
  fleetml import mlflow --model my-model --version 3
  fleetml import mlflow --model my-model --tags edge,production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		modelName, _ := cmd.Flags().GetString("model")
		version, _ := cmd.Flags().GetString("version")
		tagsStr, _ := cmd.Flags().GetString("tags")
		description, _ := cmd.Flags().GetString("description")

		if modelName == "" {
			return fmt.Errorf("--model is required")
		}

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
		}

		c := client.NewAPIClient(serverURL, apiKey)
		result, err := c.ImportMLflow(modelName, version, tags, description)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		fmt.Printf("Successfully imported model from MLflow\n")
		fmt.Printf("  Model ID:   %s\n", result["model_id"])
		fmt.Printf("  Name:       %s\n", result["name"])
		fmt.Printf("  Version:    %s\n", result["version"])
		fmt.Printf("  Format:     %s\n", result["format"])
		fmt.Printf("  Source:     %s\n", result["source"])
		if size, ok := result["artifact_size"].(float64); ok {
			fmt.Printf("  Size:       %.1f MB\n", size/1024/1024)
		}

		return nil
	},
}

var importHuggingFaceCmd = &cobra.Command{
	Use:     "huggingface",
	Short:   "Import a model from HuggingFace Hub",
	Aliases: []string{"hf"},
	Example: `  fleetml import huggingface --repo microsoft/resnet-50
  fleetml import hf --repo onnx/mobilenetv2-7 --name mobilenet --version v1
  fleetml import hf --repo my-org/private-model --filename model.onnx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		repoID, _ := cmd.Flags().GetString("repo")
		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		filename, _ := cmd.Flags().GetString("filename")
		revision, _ := cmd.Flags().GetString("revision")
		tagsStr, _ := cmd.Flags().GetString("tags")
		description, _ := cmd.Flags().GetString("description")

		if repoID == "" {
			return fmt.Errorf("--repo is required")
		}

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
		}

		c := client.NewAPIClient(serverURL, apiKey)
		result, err := c.ImportHuggingFace(repoID, name, version, filename, revision, tags, description)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		fmt.Printf("Successfully imported model from HuggingFace Hub\n")
		fmt.Printf("  Model ID:   %s\n", result["model_id"])
		fmt.Printf("  Name:       %s\n", result["name"])
		fmt.Printf("  Version:    %s\n", result["version"])
		fmt.Printf("  Format:     %s\n", result["format"])
		fmt.Printf("  Source:     %s\n", result["source"])
		if size, ok := result["artifact_size"].(float64); ok {
			fmt.Printf("  Size:       %.1f MB\n", size/1024/1024)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.AddCommand(importMLflowCmd)
	importCmd.AddCommand(importHuggingFaceCmd)

	// MLflow flags
	importMLflowCmd.Flags().String("model", "", "MLflow registered model name (required)")
	importMLflowCmd.Flags().String("version", "", "Model version (default: latest)")
	importMLflowCmd.Flags().String("tags", "", "Comma-separated tags")
	importMLflowCmd.Flags().String("description", "", "Model description")

	// HuggingFace flags
	importHuggingFaceCmd.Flags().String("repo", "", "HuggingFace repo ID e.g. microsoft/resnet-50 (required)")
	importHuggingFaceCmd.Flags().String("name", "", "Override model name (default: repo name)")
	importHuggingFaceCmd.Flags().String("version", "", "Override version (default: revision)")
	importHuggingFaceCmd.Flags().String("filename", "", "Specific file to download (default: auto-detect ONNX)")
	importHuggingFaceCmd.Flags().String("revision", "", "Git revision/branch (default: main)")
	importHuggingFaceCmd.Flags().String("tags", "", "Comma-separated tags")
	importHuggingFaceCmd.Flags().String("description", "", "Model description")
}

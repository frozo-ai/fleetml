package integrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/fleetml/fleetml/server/internal/model"
	"go.uber.org/zap"
)

// Service orchestrates importing models from external registries.
type Service struct {
	registry  *model.Registry
	mlflow    *MLflowClient
	hf        *HuggingFaceClient
	logger    *zap.SugaredLogger
}

// NewService creates a new integrations service.
func NewService(registry *model.Registry, logger *zap.SugaredLogger) *Service {
	return &Service{
		registry: registry,
		logger:   logger,
	}
}

// SetMLflowClient configures the MLflow integration.
func (s *Service) SetMLflowClient(client *MLflowClient) {
	s.mlflow = client
}

// SetHuggingFaceClient configures the HuggingFace integration.
func (s *Service) SetHuggingFaceClient(client *HuggingFaceClient) {
	s.hf = client
}

// ImportFromMLflow imports a model from an MLflow registry.
func (s *Service) ImportFromMLflow(ctx context.Context, req MLflowImportRequest) (*MLflowImportResult, error) {
	if s.mlflow == nil {
		return nil, fmt.Errorf("mlflow integration not configured (set MLFLOW_TRACKING_URI)")
	}

	// Get model version info
	version := req.Version
	if version == "" {
		// Get latest version
		m, err := s.mlflow.GetModel(ctx, req.ModelName)
		if err != nil {
			return nil, fmt.Errorf("get mlflow model: %w", err)
		}
		if len(m.LatestVersions) == 0 {
			return nil, fmt.Errorf("no versions found for model %q", req.ModelName)
		}
		version = m.LatestVersions[0].Version
	}

	mv, err := s.mlflow.GetModelVersion(ctx, req.ModelName, version)
	if err != nil {
		return nil, fmt.Errorf("get mlflow model version: %w", err)
	}

	// List artifacts to find the model file
	artifacts, err := s.mlflow.ListArtifacts(ctx, mv.RunID, "")
	if err != nil {
		return nil, fmt.Errorf("list mlflow artifacts: %w", err)
	}

	format, artifactPath := DetectFormat(artifacts)

	// Download the model artifact
	body, size, err := s.mlflow.DownloadArtifact(ctx, mv.RunID, artifactPath)
	if err != nil {
		return nil, fmt.Errorf("download mlflow artifact: %w", err)
	}
	defer body.Close()

	// Build metadata
	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["source"] = "mlflow"
	metadata["mlflow_model"] = req.ModelName
	metadata["mlflow_version"] = version
	metadata["mlflow_run_id"] = mv.RunID

	// Upload to FleetML registry
	description := req.Description
	if description == "" {
		description = mv.Description
	}

	uploadReq := model.UploadRequest{
		Name:        req.ModelName,
		Version:     "v" + version,
		Format:      format,
		Data:        body,
		Size:        size,
		Description: description,
		Metadata:    metadata,
		Tags:        req.Tags,
	}

	m, err := s.registry.Upload(ctx, uploadReq)
	if err != nil {
		return nil, fmt.Errorf("upload to registry: %w", err)
	}

	s.logger.Infow("imported model from MLflow",
		"model_id", m.ID,
		"name", m.Name,
		"version", m.Version,
		"format", format,
		"mlflow_run_id", mv.RunID,
	)

	return &MLflowImportResult{
		ModelID:      m.ID,
		Name:         m.Name,
		Version:      m.Version,
		Format:       format,
		ArtifactURL:  m.ArtifactURL,
		ArtifactSize: m.ArtifactSize,
		Source:       fmt.Sprintf("mlflow://%s/%s", req.ModelName, version),
	}, nil
}

// ImportFromHuggingFace imports a model from HuggingFace Hub.
func (s *Service) ImportFromHuggingFace(ctx context.Context, req HFImportRequest) (*HFImportResult, error) {
	if s.hf == nil {
		return nil, fmt.Errorf("huggingface integration not configured (set HF_TOKEN for private repos)")
	}

	// Get model info
	info, err := s.hf.GetModelInfo(ctx, req.RepoID)
	if err != nil {
		return nil, fmt.Errorf("get huggingface model: %w", err)
	}

	// List files to find ONNX model
	siblings, err := s.hf.ListFiles(ctx, req.RepoID, req.Revision)
	if err != nil {
		return nil, fmt.Errorf("list huggingface files: %w", err)
	}

	// Determine which file to download
	filename := req.Filename
	if filename == "" {
		onnxFile, found := FindONNXFile(siblings)
		if !found {
			format := DetectHFFormat(siblings)
			return nil, fmt.Errorf("no ONNX file found in %s (detected format: %s). Export to ONNX first or specify --filename", req.RepoID, format)
		}
		filename = onnxFile
	}

	format := "onnx"
	if hasExtension(filename, ".pt") || hasExtension(filename, ".pth") {
		format = "pytorch"
	} else if hasExtension(filename, ".tflite") {
		format = "tflite"
	}

	// Download the model file
	body, size, err := s.hf.DownloadFile(ctx, req.RepoID, filename, req.Revision)
	if err != nil {
		return nil, fmt.Errorf("download huggingface model: %w", err)
	}
	defer body.Close()

	// Derive name and version
	name := req.Name
	if name == "" {
		// Use repo name (last part of repo_id)
		parts := strings.Split(req.RepoID, "/")
		name = parts[len(parts)-1]
	}
	version := req.Version
	if version == "" {
		revision := req.Revision
		if revision == "" {
			revision = "main"
		}
		version = revision
	}

	// Build metadata
	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["source"] = "huggingface"
	metadata["hf_repo_id"] = req.RepoID
	metadata["hf_filename"] = filename
	metadata["hf_pipeline"] = info.Pipeline
	metadata["hf_library"] = info.LibraryName
	metadata["hf_downloads"] = info.Downloads

	// Build tags
	tags := req.Tags
	if len(tags) == 0 {
		tags = info.Tags
	}

	description := req.Description
	if description == "" {
		description = fmt.Sprintf("Imported from HuggingFace Hub: %s", req.RepoID)
	}

	uploadReq := model.UploadRequest{
		Name:        name,
		Version:     version,
		Format:      format,
		Data:        body,
		Size:        size,
		Description: description,
		Metadata:    metadata,
		Tags:        tags,
	}

	m, err := s.registry.Upload(ctx, uploadReq)
	if err != nil {
		return nil, fmt.Errorf("upload to registry: %w", err)
	}

	s.logger.Infow("imported model from HuggingFace",
		"model_id", m.ID,
		"name", m.Name,
		"version", m.Version,
		"repo_id", req.RepoID,
		"filename", filename,
	)

	return &HFImportResult{
		ModelID:      m.ID,
		Name:         m.Name,
		Version:      m.Version,
		Format:       format,
		ArtifactURL:  m.ArtifactURL,
		ArtifactSize: m.ArtifactSize,
		Source:       fmt.Sprintf("hf://%s/%s", req.RepoID, filename),
	}, nil
}

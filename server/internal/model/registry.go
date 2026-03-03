package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/fleetml/fleetml/server/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Registry manages the model registry.
type Registry struct {
	db      *pgxpool.Pool
	storage storage.ObjectStore
}

func NewRegistry(db *pgxpool.Pool, store storage.ObjectStore) *Registry {
	return &Registry{db: db, storage: store}
}

// UploadRequest contains the data needed to upload a model.
type UploadRequest struct {
	Name        string
	Version     string
	Format      string
	Data        io.Reader
	Size        int64
	Description string
	Metadata    map[string]interface{}
	Tags        []string
	CreatedBy   *string
}

// Upload uploads a model file and creates a registry entry.
func (r *Registry) Upload(ctx context.Context, req UploadRequest) (*domain.Model, error) {
	// Compute checksum while uploading
	h := sha256.New()
	teeReader := io.TeeReader(req.Data, h)

	// Build S3 key
	key := fmt.Sprintf("%s/%s/model.%s", req.Name, req.Version, req.Format)
	artifactURL := fmt.Sprintf("s3://fleetml-models/%s", key)

	if err := r.storage.Upload(ctx, key, teeReader, req.Size, "application/octet-stream"); err != nil {
		return nil, fmt.Errorf("upload model to storage: %w", err)
	}

	checksum := "sha256:" + hex.EncodeToString(h.Sum(nil))

	metadataJSON, _ := json.Marshal(req.Metadata)

	var m domain.Model
	err := r.db.QueryRow(ctx, `
		INSERT INTO models (name, version, format, artifact_url, artifact_size, checksum, description, metadata, tags, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, version, format, artifact_url, artifact_size, checksum, description, created_at`,
		req.Name, req.Version, req.Format, artifactURL, req.Size, checksum,
		req.Description, metadataJSON, req.Tags, req.CreatedBy,
	).Scan(&m.ID, &m.Name, &m.Version, &m.Format, &m.ArtifactURL, &m.ArtifactSize,
		&m.Checksum, &m.Description, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert model: %w", err)
	}

	m.Metadata = req.Metadata
	m.Tags = req.Tags
	return &m, nil
}

// GetModel returns a model by name and version.
func (r *Registry) GetModel(ctx context.Context, name, version string) (*domain.Model, error) {
	var m domain.Model
	var metadataJSON, variantsJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, name, version, format, artifact_url, artifact_size, checksum,
			   description, metadata, tags, compiled_variants, created_at, created_by
		FROM models WHERE name = $1 AND version = $2`, name, version,
	).Scan(
		&m.ID, &m.Name, &m.Version, &m.Format, &m.ArtifactURL, &m.ArtifactSize,
		&m.Checksum, &m.Description, &metadataJSON, &m.Tags, &variantsJSON,
		&m.CreatedAt, &m.CreatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("model %s:%s not found", name, version)
		}
		return nil, fmt.Errorf("get model: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &m.Metadata)
	}
	if variantsJSON != nil {
		json.Unmarshal(variantsJSON, &m.CompiledVariants)
	}

	return &m, nil
}

// GetModelByID returns a model by its UUID.
func (r *Registry) GetModelByID(ctx context.Context, id string) (*domain.Model, error) {
	var m domain.Model
	var metadataJSON, variantsJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, name, version, format, artifact_url, artifact_size, checksum,
			   description, metadata, tags, compiled_variants, created_at, created_by
		FROM models WHERE id = $1`, id,
	).Scan(
		&m.ID, &m.Name, &m.Version, &m.Format, &m.ArtifactURL, &m.ArtifactSize,
		&m.Checksum, &m.Description, &metadataJSON, &m.Tags, &variantsJSON,
		&m.CreatedAt, &m.CreatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("model %s not found", id)
		}
		return nil, fmt.Errorf("get model: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &m.Metadata)
	}
	if variantsJSON != nil {
		json.Unmarshal(variantsJSON, &m.CompiledVariants)
	}

	return &m, nil
}

// ListModels lists models with optional filters.
func (r *Registry) ListModels(ctx context.Context, filter domain.ModelFilter) ([]*domain.Model, int, error) {
	query := `SELECT id, name, version, format, artifact_url, artifact_size, checksum,
			         description, metadata, tags, created_at
			  FROM models WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM models WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Name != "" {
		query += fmt.Sprintf(" AND name = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND name = $%d", argIdx)
		args = append(args, filter.Name)
		argIdx++
	}

	var total int
	r.db.QueryRow(ctx, countQuery, args...).Scan(&total)

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()

	var models []*domain.Model
	for rows.Next() {
		var m domain.Model
		var metadataJSON []byte
		err := rows.Scan(
			&m.ID, &m.Name, &m.Version, &m.Format, &m.ArtifactURL, &m.ArtifactSize,
			&m.Checksum, &m.Description, &metadataJSON, &m.Tags, &m.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan model: %w", err)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &m.Metadata)
		}
		models = append(models, &m)
	}

	return models, total, nil
}

// DeleteModel soft-deletes a model (cannot delete if actively deployed).
func (r *Registry) DeleteModel(ctx context.Context, id string) error {
	// Check if model is actively deployed
	var activeCount int
	r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM device_models WHERE model_id = $1 AND status = 'active'`, id,
	).Scan(&activeCount)

	if activeCount > 0 {
		return fmt.Errorf("cannot delete model: still deployed to %d devices", activeCount)
	}

	_, err := r.db.Exec(ctx, `DELETE FROM models WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete model: %w", err)
	}
	return nil
}

// AddCompiledVariant appends a compiled variant to the model's compiled_variants JSONB.
func (r *Registry) AddCompiledVariant(ctx context.Context, modelID string, variant domain.CompiledVariant) error {
	// Get current variants
	m, err := r.GetModelByID(ctx, modelID)
	if err != nil {
		return fmt.Errorf("get model: %w", err)
	}

	// Replace existing variant for this runtime, or append new one
	variants := m.CompiledVariants
	found := false
	for i, v := range variants {
		if v.Runtime == variant.Runtime {
			variants[i] = variant
			found = true
			break
		}
	}
	if !found {
		variants = append(variants, variant)
	}

	variantsJSON, err := json.Marshal(variants)
	if err != nil {
		return fmt.Errorf("marshal variants: %w", err)
	}

	_, err = r.db.Exec(ctx, `UPDATE models SET compiled_variants = $1 WHERE id = $2`, variantsJSON, modelID)
	if err != nil {
		return fmt.Errorf("update compiled variants: %w", err)
	}

	return nil
}

// GetVariantForRuntime returns the compiled variant matching the given runtime, or nil if none exists.
func (r *Registry) GetVariantForRuntime(ctx context.Context, modelID, runtime string) (*domain.CompiledVariant, error) {
	m, err := r.GetModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	for _, v := range m.CompiledVariants {
		if v.Runtime == runtime {
			return &v, nil
		}
	}

	return nil, nil
}

// GetArtifactURL generates a pre-signed URL for model download.
func (r *Registry) GetArtifactURL(ctx context.Context, modelID string) (string, error) {
	m, err := r.GetModelByID(ctx, modelID)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s/%s/model.%s", m.Name, m.Version, m.Format)
	url, err := r.storage.GeneratePresignedURL(ctx, key, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("generate presigned url: %w", err)
	}

	return url, nil
}

// GetVariantArtifactURL generates a pre-signed URL for a compiled variant's artifact.
func (r *Registry) GetVariantArtifactURL(ctx context.Context, modelID, runtime string) (string, error) {
	variant, err := r.GetVariantForRuntime(ctx, modelID, runtime)
	if err != nil {
		return "", err
	}
	if variant == nil {
		return "", fmt.Errorf("no variant for runtime %s", runtime)
	}

	// The variant's ArtifactURL is like s3://bucket/model-id/compiled/runtime/model.ext
	// Extract the key portion after s3://bucket/
	key := extractS3Key(variant.ArtifactURL)
	url, err := r.storage.GeneratePresignedURL(ctx, key, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("generate variant presigned url: %w", err)
	}

	return url, nil
}

// extractS3Key extracts the object key from an s3:// URL.
func extractS3Key(s3URL string) string {
	// s3://bucket-name/path/to/key -> path/to/key
	const prefix = "s3://"
	if len(s3URL) <= len(prefix) {
		return s3URL
	}
	rest := s3URL[len(prefix):]
	// Find first slash after bucket name
	for i, c := range rest {
		if c == '/' {
			return rest[i+1:]
		}
	}
	return rest
}

package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/fleetml/fleetml/server/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ModelHandler handles model-related endpoints.
type ModelHandler struct {
	registry *model.Registry
	logger   *zap.SugaredLogger
}

func NewModelHandler(registry *model.Registry, logger *zap.SugaredLogger) *ModelHandler {
	return &ModelHandler{registry: registry, logger: logger}
}

// Upload handles multipart model upload.
func (h *ModelHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 500MB)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		http.Error(w, `{"error":"failed to parse multipart form"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	version := r.FormValue("version")
	format := r.FormValue("format")

	if name == "" || version == "" || format == "" {
		http.Error(w, `{"error":"name, version, and format are required"}`, http.StatusBadRequest)
		return
	}

	var metadata map[string]interface{}
	if metaStr := r.FormValue("metadata"); metaStr != "" {
		json.Unmarshal([]byte(metaStr), &metadata)
	}

	var tags []string
	if tagsStr := r.FormValue("tags"); tagsStr != "" {
		// Split comma-separated tags
		for _, t := range splitTags(tagsStr) {
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	m, err := h.registry.Upload(r.Context(), model.UploadRequest{
		Name:        name,
		Version:     version,
		Format:      format,
		Data:        file,
		Size:        header.Size,
		Description: r.FormValue("description"),
		Metadata:    metadata,
		Tags:        tags,
		OrgID:       claims.OrgID,
	})
	if err != nil {
		h.logger.Errorw("failed to upload model", "error", err)
		http.Error(w, `{"error":"failed to upload model"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// List lists models with optional filters.
func (h *ModelHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	filter := domain.ModelFilter{
		Name: r.URL.Query().Get("name"),
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		filter.Limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		filter.Offset, _ = strconv.Atoi(o)
	}

	models, total, err := h.registry.ListModels(r.Context(), claims.OrgID, filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list models"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"models": models,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Get returns a single model by ID.
func (h *ModelHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	m, err := h.registry.GetModelByID(r.Context(), id, claims.OrgID)
	if err != nil {
		http.Error(w, `{"error":"model not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// Delete deletes a model.
func (h *ModelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.registry.DeleteModel(r.Context(), claims.OrgID, id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func splitTags(s string) []string {
	var tags []string
	current := ""
	for _, c := range s {
		if c == ',' {
			tags = append(tags, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		tags = append(tags, current)
	}
	return tags
}

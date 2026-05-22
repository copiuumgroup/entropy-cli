package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/copiuumgroup/entropy-cli/internal/config"
	"github.com/copiuumgroup/entropy-cli/internal/database"
	"github.com/gorilla/mux"
)

const defaultPageSize = 20

// HealthHandler handles the health check endpoint.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListDownloadsHandler lists all downloads with pagination.
func ListDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	page := getPageParam(r, "page", 1)
	pageSize := getPageParam(r, "page_size", defaultPageSize)

	downloads, total, err := database.ListDownloads(page, pageSize)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to list downloads", err.Error())
		return
	}

	responses := make([]DownloadResponse, len(downloads))
	for i, d := range downloads {
		responses[i] = fromDownload(&d)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	paginatedResponse(w, responses, total, page, pageSize, totalPages)
}

// GetDownloadHandler retrieves a specific download.
func GetDownloadHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid download ID", "")
		return
	}

	download, err := database.GetDownload(uint(id))
	if err != nil {
		errorResponse(w, http.StatusNotFound, "Download not found", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fromDownload(download))
}

// CreateDownloadHandler creates a new download.
func CreateDownloadHandler(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	if req.URL == "" {
		errorResponse(w, http.StatusBadRequest, "URL is required", "")
		return
	}

	format := req.Format
	if format == "" {
		format = config.C.Quality
	}

	download := &database.Download{
		URL:       req.URL,
		Status:    "queued",
		Progress:  0,
		Format:    format,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.CreateDownload(download); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to create download", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fromDownload(download))
}

// UpdateDownloadHandler updates a download.
func UpdateDownloadHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid download ID", "")
		return
	}

	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	download, err := database.GetDownload(uint(id))
	if err != nil {
		errorResponse(w, http.StatusNotFound, "Download not found", "")
		return
	}

	// Update fields from request
	if status, ok := update["status"].(string); ok {
		download.Status = status
	}
	if progress, ok := update["progress"].(float64); ok {
		download.Progress = progress
	}
	if speed, ok := update["speed"].(string); ok {
		download.Speed = speed
	}
	if title, ok := update["title"].(string); ok {
		download.Title = title
	}
	if uploader, ok := update["uploader"].(string); ok {
		download.Uploader = uploader
	}

	download.UpdatedAt = time.Now()

	if err := database.UpdateDownload(download); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to update download", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fromDownload(download))
}

// DeleteDownloadHandler deletes a download.
func DeleteDownloadHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid download ID", "")
		return
	}

	if err := database.DeleteDownload(uint(id)); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to delete download", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListDownloadsByStatusHandler lists downloads by status.
func ListDownloadsByStatusHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	status := params["status"]

	page := getPageParam(r, "page", 1)
	pageSize := getPageParam(r, "page_size", defaultPageSize)

	downloads, total, err := database.ListDownloadsByStatus(status, page, pageSize)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to list downloads", err.Error())
		return
	}

	responses := make([]DownloadResponse, len(downloads))
	for i, d := range downloads {
		responses[i] = fromDownload(&d)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	paginatedResponse(w, responses, total, page, pageSize, totalPages)
}

// ListLibraryHandler lists library items.
func ListLibraryHandler(w http.ResponseWriter, r *http.Request) {
	page := getPageParam(r, "page", 1)
	pageSize := getPageParam(r, "page_size", defaultPageSize)

	items, total, err := database.ListLibraryItems(page, pageSize)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to list library", err.Error())
		return
	}

	responses := make([]LibraryItemResponse, len(items))
	for i, item := range items {
		responses[i] = fromLibraryItem(&item)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	paginatedResponse(w, responses, total, page, pageSize, totalPages)
}

// SearchLibraryHandler searches the library.
func SearchLibraryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		errorResponse(w, http.StatusBadRequest, "Query parameter 'q' is required", "")
		return
	}

	page := getPageParam(r, "page", 1)
	pageSize := getPageParam(r, "page_size", defaultPageSize)

	items, total, err := database.SearchLibrary(query, page, pageSize)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to search library", err.Error())
		return
	}

	responses := make([]LibraryItemResponse, len(items))
	for i, item := range items {
		responses[i] = fromLibraryItem(&item)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	paginatedResponse(w, responses, total, page, pageSize, totalPages)
}

// DeleteLibraryItemHandler deletes a library item.
func DeleteLibraryItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid library item ID", "")
		return
	}

	if err := database.DeleteLibraryItem(uint(id)); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to delete library item", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetSettingsHandler returns current settings.
func GetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"output_dir":     config.C.OutputDir,
		"quality":        config.C.Quality,
		"connections":    config.C.Connections,
		"splits":         config.C.Splits,
		"max_concurrent": config.C.MaxConcurrent,
	})
}

// UpdateSettingsHandler updates settings.
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var req SettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	if req.OutputDir != "" {
		config.C.OutputDir = req.OutputDir
	}
	if req.Quality != "" {
		config.C.Quality = req.Quality
	}
	if req.Connections > 0 {
		config.C.Connections = req.Connections
	}
	if req.Splits > 0 {
		config.C.Splits = req.Splits
	}
	if req.MaxConcurrent > 0 {
		config.C.MaxConcurrent = req.MaxConcurrent
	}

	if err := config.Save(); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to save settings", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"output_dir":     config.C.OutputDir,
		"quality":        config.C.Quality,
		"connections":    config.C.Connections,
		"splits":         config.C.Splits,
		"max_concurrent": config.C.MaxConcurrent,
	})
}

// Helper functions

func getPageParam(r *http.Request, param string, defaultValue int) int {
	val := r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}
	page, err := strconv.Atoi(val)
	if err != nil || page < 1 {
		return defaultValue
	}
	return page
}

func errorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Details: details,
	})
}

func paginatedResponse(w http.ResponseWriter, data interface{}, total int64, page, pageSize int, totalPages int64) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

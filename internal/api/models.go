package api

import (
	"time"

	"github.com/copiuumgroup/entropy-cli/internal/database"
)

// DownloadRequest is the request body for creating a download.
type DownloadRequest struct {
	URL    string `json:"url" binding:"required"`
	Format string `json:"format,omitempty"`
}

// DownloadResponse is the response body for download operations.
type DownloadResponse struct {
	ID          uint       `json:"id"`
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Uploader    string     `json:"uploader"`
	Status      string     `json:"status"`
	Progress    float64    `json:"progress"`
	Speed       string     `json:"speed,omitempty"`
	Error       string     `json:"error,omitempty"`
	FilePath    string     `json:"file_path,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// fromDownload converts a database Download to an API response.
func fromDownload(d *database.Download) DownloadResponse {
	return DownloadResponse{
		ID:          d.ID,
		URL:         d.URL,
		Title:       d.Title,
		Uploader:    d.Uploader,
		Status:      d.Status,
		Progress:    d.Progress,
		Speed:       d.Speed,
		Error:       d.Error,
		FilePath:    d.FilePath,
		CreatedAt:   d.CreatedAt,
		CompletedAt: d.CompletedAt,
	}
}

// LibraryItemResponse is the response for library items.
type LibraryItemResponse struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
	Size     int64  `json:"size"`
	Format   string `json:"format"`
	AddedAt  time.Time `json:"added_at"`
}

func fromLibraryItem(item *database.LibraryItem) LibraryItemResponse {
	return LibraryItemResponse{
		ID:       item.ID,
		Title:    item.Title,
		Artist:   item.Artist,
		Album:    item.Album,
		Duration: item.Duration,
		Size:     item.Size,
		Format:   item.Format,
		AddedAt:  item.AddedAt,
	}
}

// ProgressUpdate is sent over WebSocket for real-time progress.
type ProgressUpdate struct {
	DownloadID uint    `json:"download_id"`
	Progress   float64 `json:"progress"`
	Speed      string  `json:"speed"`
	Status     string  `json:"status"`
	Error      string  `json:"error,omitempty"`
}

// SettingsRequest is the request for updating settings.
type SettingsRequest struct {
	OutputDir     string `json:"output_dir,omitempty"`
	Quality       string `json:"quality,omitempty"`
	Connections   int    `json:"connections,omitempty"`
	Splits        int    `json:"splits,omitempty"`
	MaxConcurrent int    `json:"max_concurrent,omitempty"`
}

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse wraps paginated results.
type PaginatedResponse struct {
	Data      interface{} `json:"data"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	TotalPages int64      `json:"total_pages"`
}

package database

import (
	"time"

	"gorm.io/gorm"
)

// Download represents a single download task in the system.
type Download struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Uploader     string    `json:"uploader"`
	Duration     int       `json:"duration"` // in seconds
	Status       string    `json:"status"` // "queued", "downloading", "processing", "done", "error", "cancelled"
	Progress     float64   `json:"progress"` // 0.0 to 100.0
	Size         int64     `json:"size"` // bytes
	Speed        string    `json:"speed"` // e.g., "5.2 MB/s"
	Error        string    `json:"error,omitempty"`
	FilePath     string    `json:"file_path,omitempty"`
	Format       string    `json:"format"` // "mp3", "flac", etc.
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LibraryItem represents a file in the user's music library.
type LibraryItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FilePath     string    `gorm:"uniqueIndex" json:"file_path"`
	FileName     string    `json:"file_name"`
	Title        string    `json:"title"`
	Artist       string    `json:"artist"`
	Album        string    `json:"album"`
	Genre        string    `json:"genre,omitempty"`
	Duration     int       `json:"duration"` // seconds
	Bitrate      int       `json:"bitrate"` // kbps
	Size         int64     `json:"size"` // bytes
	Format       string    `json:"format"` // "mp3", "flac", etc.
	DownloadID   *uint     `json:"download_id,omitempty"` // reference to Download if imported via entropy
	AddedAt      time.Time `json:"added_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SearchCache caches search results to avoid redundant yt-dlp calls.
type SearchCache struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Query     string    `gorm:"index" json:"query"`
	Provider  string    `json:"provider"` // "youtube", "soundcloud"
	Results   string    `json:"results"` // JSON string of results
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for GORM.
func (SearchCache) TableName() string {
	return "search_cache"
}

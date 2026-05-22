package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init initializes the database connection and runs migrations.
func Init() error {
	dbPath := dbPath()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	DB = db
	return migrate()
}

// dbPath returns the path to the SQLite database file.
func dbPath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "entropy-cli", "entropy.db")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "entropy-cli", "entropy.db")
}

// migrate runs all database migrations.
func migrate() error {
	return DB.AutoMigrate(
		&Download{},
		&LibraryItem{},
		&SearchCache{},
	)
}

// GetDownload retrieves a download by ID.
func GetDownload(id uint) (*Download, error) {
	var download Download
	if err := DB.First(&download, id).Error; err != nil {
		return nil, err
	}
	return &download, nil
}

// ListDownloads retrieves downloads with pagination.
func ListDownloads(page, pageSize int) ([]Download, int64, error) {
	var downloads []Download
	var total int64

	if err := DB.Model(&Download{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&downloads).Error; err != nil {
		return nil, 0, err
	}

	return downloads, total, nil
}

// ListDownloadsByStatus retrieves downloads filtered by status.
func ListDownloadsByStatus(status string, page, pageSize int) ([]Download, int64, error) {
	var downloads []Download
	var total int64

	if err := DB.Model(&Download{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Where("status = ?", status).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&downloads).Error; err != nil {
		return nil, 0, err
	}

	return downloads, total, nil
}

// CreateDownload creates a new download record.
func CreateDownload(download *Download) error {
	return DB.Create(download).Error
}

// UpdateDownload updates an existing download record.
func UpdateDownload(download *Download) error {
	return DB.Save(download).Error
}

// DeleteDownload deletes a download record.
func DeleteDownload(id uint) error {
	return DB.Delete(&Download{}, id).Error
}

// GetLibraryItem retrieves a library item by ID.
func GetLibraryItem(id uint) (*LibraryItem, error) {
	var item LibraryItem
	if err := DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// ListLibraryItems retrieves library items with pagination.
func ListLibraryItems(page, pageSize int) ([]LibraryItem, int64, error) {
	var items []LibraryItem
	var total int64

	if err := DB.Model(&LibraryItem{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Offset(offset).Limit(pageSize).Order("added_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// SearchLibrary searches library items by title or artist.
func SearchLibrary(query string, page, pageSize int) ([]LibraryItem, int64, error) {
	var items []LibraryItem
	var total int64

	search := "SELECT * FROM library_items WHERE title LIKE ? OR artist LIKE ?"
	pattern := "%" + query + "%"

	if err := DB.Model(&LibraryItem{}).
		Where("title LIKE ? OR artist LIKE ?", pattern, pattern).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Where("title LIKE ? OR artist LIKE ?", pattern, pattern).
		Offset(offset).
		Limit(pageSize).
		Order("title ASC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CreateLibraryItem creates a new library item.
func CreateLibraryItem(item *LibraryItem) error {
	return DB.Create(item).Error
}

// DeleteLibraryItem deletes a library item.
func DeleteLibraryItem(id uint) error {
	return DB.Delete(&LibraryItem{}, id).Error
}

// GetSearchCache retrieves cached search results.
func GetSearchCache(query, provider string) (*SearchCache, error) {
	var cache SearchCache
	now := time.Now()

	if err := DB.Where("query = ? AND provider = ? AND expires_at > ?", query, provider, now).
		First(&cache).Error; err != nil {
		return nil, err
	}

	return &cache, nil
}

// CreateSearchCache stores search results in cache.
func CreateSearchCache(cache *SearchCache) error {
	cache.ExpiresAt = time.Now().Add(24 * time.Hour) // 24 hour TTL
	return DB.Create(cache).Error
}

// CleanExpiredCache removes expired search cache entries.
func CleanExpiredCache() error {
	return DB.Where("expires_at < ?", time.Now()).Delete(&SearchCache{}).Error
}

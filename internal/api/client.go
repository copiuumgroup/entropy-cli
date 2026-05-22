package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Client is an HTTP client for the entropy API.
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// GetHealth checks if the API is healthy.
func (c *Client) GetHealth() error {
	resp, err := c.client.Get(c.baseURL + "/api/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

// ListDownloads retrieves all downloads.
func (c *Client) ListDownloads(page, pageSize int) ([]DownloadResponse, error) {
	url := fmt.Sprintf("%s/api/downloads?page=%d&page_size=%d", c.baseURL, page, pageSize)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list downloads: %d", resp.StatusCode)
	}

	var paginated PaginatedResponse
	if err := json.NewDecoder(resp.Body).Decode(&paginated); err != nil {
		return nil, err
	}

	// Convert interface{} to []DownloadResponse
	var downloads []DownloadResponse
	if data, ok := paginated.Data.([]interface{}); ok {
		for _, d := range data {
			if bz, err := json.Marshal(d); err == nil {
				var dl DownloadResponse
				if err := json.Unmarshal(bz, &dl); err == nil {
					downloads = append(downloads, dl)
				}
			}
		}
	}

	return downloads, nil
}

// CreateDownload creates a new download.
func (c *Client) CreateDownload(req DownloadRequest) (*DownloadResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/api/downloads", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create download: %d", resp.StatusCode)
	}

	var download DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&download); err != nil {
		return nil, err
	}

	return &download, nil
}

// GetDownload retrieves a specific download.
func (c *Client) GetDownload(id uint) (*DownloadResponse, error) {
	url := fmt.Sprintf("%s/api/downloads/%d", c.baseURL, id)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download not found: %d", resp.StatusCode)
	}

	var download DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&download); err != nil {
		return nil, err
	}

	return &download, nil
}

// UpdateDownload updates a download.
func (c *Client) UpdateDownload(id uint, update map[string]interface{}) (*DownloadResponse, error) {
	body, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/downloads/%d", c.baseURL, id)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update download: %d", resp.StatusCode)
	}

	var download DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&download); err != nil {
		return nil, err
	}

	return &download, nil
}

// DeleteDownload deletes a download.
func (c *Client) DeleteDownload(id uint) error {
	url := fmt.Sprintf("%s/api/downloads/%d", c.baseURL, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete download: %d", resp.StatusCode)
	}

	return nil
}

// ListLibrary retrieves library items.
func (c *Client) ListLibrary(page, pageSize int) ([]LibraryItemResponse, error) {
	url := fmt.Sprintf("%s/api/library?page=%d&page_size=%d", c.baseURL, page, pageSize)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list library: %d", resp.StatusCode)
	}

	var paginated PaginatedResponse
	if err := json.NewDecoder(resp.Body).Decode(&paginated); err != nil {
		return nil, err
	}

	var items []LibraryItemResponse
	if data, ok := paginated.Data.([]interface{}); ok {
		for _, d := range data {
			if bz, err := json.Marshal(d); err == nil {
				var item LibraryItemResponse
				if err := json.Unmarshal(bz, &item); err == nil {
					items = append(items, item)
				}
			}
		}
	}

	return items, nil
}

// SearchLibrary searches library items.
func (c *Client) SearchLibrary(query string, page, pageSize int) ([]LibraryItemResponse, error) {
	url := fmt.Sprintf("%s/api/library/search?q=%s&page=%d&page_size=%d", c.baseURL, query, page, pageSize)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search library: %d", resp.StatusCode)
	}

	var paginated PaginatedResponse
	if err := json.NewDecoder(resp.Body).Decode(&paginated); err != nil {
		return nil, err
	}

	var items []LibraryItemResponse
	if data, ok := paginated.Data.([]interface{}); ok {
		for _, d := range data {
			if bz, err := json.Marshal(d); err == nil {
				var item LibraryItemResponse
				if err := json.Unmarshal(bz, &item); err == nil {
					items = append(items, item)
				}
			}
		}
	}

	return items, nil
}

// GetSettings retrieves current settings.
func (c *Client) GetSettings() (map[string]interface{}, error) {
	resp, err := c.client.Get(c.baseURL + "/api/settings")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get settings: %d", resp.StatusCode)
	}

	var settings map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateSettings updates settings.
func (c *Client) UpdateSettings(req SettingsRequest) (map[string]interface{}, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	req_http, err := http.NewRequest(http.MethodPut, c.baseURL+"/api/settings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req_http.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req_http)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update settings: %d", resp.StatusCode)
	}

	var settings map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// readAllWithLimit reads up to limit bytes from r.
func readAllWithLimit(r io.Reader, limit int64) ([]byte, error) {
	lr := io.LimitReader(r, limit)
	return io.ReadAll(lr)
}

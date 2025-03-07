package harvest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"harvest-cli/pkg/config"
)

// Client represents a Harvest API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	accountID  string
	token      string
}

// TimeEntry represents a time entry in Harvest
type TimeEntry struct {
	ID             int64          `json:"id,omitempty"`
	SpentDate      string         `json:"spent_date"`
	ProjectID      int            `json:"project_id"`
	TaskID         int            `json:"task_id"`
	Hours          float64        `json:"hours"`
	Notes          string         `json:"notes,omitempty"`
	CreatedAt      time.Time      `json:"created_at,omitempty"`
	UpdatedAt      time.Time      `json:"updated_at,omitempty"`
	IsRunning      bool           `json:"is_running,omitempty"`
	User           User           `json:"user,omitempty"`
	UserID         int64          `json:"user_id,omitempty"`
	UserAssignment UserAssignment `json:"user_assignment,omitempty"`
}

// User represents a user in Harvest
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// UserAssignment represents a user assignment in Harvest
type UserAssignment struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}

// ErrorResponse represents an error response from the Harvest API
type ErrorResponse struct {
	Message string `json:"message"`
}

// NewClient creates a new Harvest API client
func NewClient(cfg *config.APIConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:   cfg.BaseURL,
		accountID: cfg.AccountID,
		token:     cfg.Token,
	}
}

// CreateTimeEntry creates a new time entry in Harvest
func (c *Client) CreateTimeEntry(entry *TimeEntry) (*TimeEntry, error) {
	url := fmt.Sprintf("%s/time_entries", c.baseURL)

	body, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal time entry: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Harvest-Account-ID", c.accountID)
	req.Header.Set("User-Agent", "Harvest CLI Utility")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("failed to parse error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s (status code: %d)", errResp.Message, resp.StatusCode)
	}

	// Parse response
	var timeEntry TimeEntry
	if err := json.Unmarshal(respBody, &timeEntry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &timeEntry, nil
}

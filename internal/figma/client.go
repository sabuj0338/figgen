package figma

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Client for interacting with Figma API
type Client struct {
	HTTPClient *http.Client
	Token      string
}

// NewClient initializes a new Figma API client using the FIGMA_TOKEN env var
func NewClient() (*Client, error) {
	token := os.Getenv("FIGMA_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("FIGMA_TOKEN environment variable is required")
	}

	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Token:      token,
	}, nil
}

// ParseURL extracts the file key and optional node ID from a Figma URL.
func ParseURL(figmaURL string) (fileKey string, nodeID string, err error) {
	re := regexp.MustCompile(`figma\.com/(?:file|design)/([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(figmaURL)
	if len(matches) < 2 {
		return "", "", fmt.Errorf("could not extract file key from URL")
	}
	fileKey = matches[1]

	if strings.Contains(figmaURL, "node-id=") {
		parts := strings.Split(figmaURL, "node-id=")
		if len(parts) == 2 {
			nodeID = strings.Split(parts[1], "&")[0]
			nodeID = strings.ReplaceAll(nodeID, "%3A", ":") // decode url-encoded colons
		}
	}

	return fileKey, nodeID, nil
}

// FetchFile fetches the Figma document
func (c *Client) FetchFile(fileKey string, nodeID string) (*FileResponse, error) {
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey)
	if nodeID != "" {
		url = fmt.Sprintf("%s?ids=%s", url, nodeID)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("X-Figma-Token", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from figma api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("figma api returned status %d", resp.StatusCode)
	}

	var fileResp FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		return nil, fmt.Errorf("failed to decode figma response: %w", err)
	}

	return &fileResp, nil
}

package figma

import (
	"encoding/json"
	"fmt"
	"io"
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
			nodeID = strings.ReplaceAll(nodeID, "-", ":")   // decode dashes used in URLs
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
	if nodeID != "" {
		var nodeResp struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Nodes   map[string]struct {
				Document Document `json:"document"`
			} `json:"nodes"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&nodeResp); err != nil {
			return nil, fmt.Errorf("failed to decode figma node response: %w", err)
		}

		if nodeData, ok := nodeResp.Nodes[nodeID]; ok {
			fileResp.Name = nodeResp.Name
			fileResp.Version = nodeResp.Version
			fileResp.Document = nodeData.Document
		} else {
			return nil, fmt.Errorf("node %s not found in response", nodeID)
		}
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
			return nil, fmt.Errorf("failed to decode figma response: %w", err)
		}
	}

	return &fileResp, nil
}

// ImageResponse represents the response from the Figma images API
type ImageResponse struct {
	Err    string            `json:"err"`
	Images map[string]string `json:"images"`
}

// FetchImageURLs fetches the render URLs for specific nodes
func (c *Client) FetchImageURLs(fileKey string, nodeIDs []string) (map[string]string, error) {
	if len(nodeIDs) == 0 {
		return make(map[string]string), nil
	}

	url := fmt.Sprintf("https://api.figma.com/v1/images/%s?ids=%s&format=png", fileKey, strings.Join(nodeIDs, ","))
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

	var imageResp ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		return nil, fmt.Errorf("failed to decode figma response: %w", err)
	}

	if imageResp.Err != "" && imageResp.Err != "null" {
		return nil, fmt.Errorf("figma images api error: %s", imageResp.Err)
	}

	return imageResp.Images, nil
}

// DownloadImages downloads a map of image URLs to the specified directory
func (c *Client) DownloadImages(images map[string]string, destDir string) error {
	if len(images) == 0 {
		return nil
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	for id, imgURL := range images {
		if imgURL == "" {
			continue
		}

		resp, err := http.Get(imgURL)
		if err != nil {
			fmt.Printf("Warning: failed to download image %s: %v\n", id, err)
			continue
		}
		
		// Clean the ID for filename (e.g. "0:1" -> "0_1")
		safeID := strings.ReplaceAll(id, ":", "_")
		filePath := fmt.Sprintf("%s/%s.png", destDir, safeID)
		
		file, err := os.Create(filePath)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}

		_, err = io.Copy(file, resp.Body)
		file.Close()
		resp.Body.Close()

		if err != nil {
			return fmt.Errorf("failed to save image %s: %w", filePath, err)
		}
	}

	return nil
}

package figma

// FileResponse represents the top-level JSON response from the Figma API
type FileResponse struct {
	Document Document `json:"document"`
	Name     string   `json:"name"`
	Version  string   `json:"version"`
}

// Document is the root node of the Figma file
type Document struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Children []Node `json:"children"`
}

// Node represents any element in the Figma document (Frame, Component, Text, etc.)
type Node struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Children []Node                 `json:"children,omitempty"`
	Style    map[string]interface{} `json:"style,omitempty"`
}

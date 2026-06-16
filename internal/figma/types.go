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
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Type                  string                 `json:"type"`
	Characters            string                 `json:"characters,omitempty"`
	LayoutMode            string                 `json:"layoutMode,omitempty"`
	PrimaryAxisAlignItems string                 `json:"primaryAxisAlignItems,omitempty"`
	CounterAxisAlignItems string                 `json:"counterAxisAlignItems,omitempty"`
	ItemSpacing           float64                `json:"itemSpacing,omitempty"`
	PaddingLeft           float64                `json:"paddingLeft,omitempty"`
	PaddingTop            float64                `json:"paddingTop,omitempty"`
	PaddingRight          float64                `json:"paddingRight,omitempty"`
	PaddingBottom         float64                `json:"paddingBottom,omitempty"`
	Children              []Node                 `json:"children,omitempty"`
	Style                 map[string]interface{} `json:"style,omitempty"`
	Fills                 []Fill                 `json:"fills,omitempty"`
}

// Fill represents a fill on a node
type Fill struct {
	Type     string `json:"type"`
	ImageRef string `json:"imageRef,omitempty"`
}

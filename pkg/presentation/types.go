package presentation

import "time"

const (
	SlideWidth  = 1920
	SlideHeight = 1080
)

// Presentation represents an HTML slide presentation
type Presentation struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	SlideCount int       `json:"slide_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Metadata is stored in metadata.json
type Metadata struct {
	Name       string    `json:"name"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	SlideCount int       `json:"slide_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Slide represents a single slide's content
type Slide struct {
	Number      int    `json:"number"`
	HTMLContent string `json:"html_content"`
}

// PresentationInfo is a lightweight summary for listing
type PresentationInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	SlideCount int       `json:"slide_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	FilePath   string    `json:"file_path"`
}

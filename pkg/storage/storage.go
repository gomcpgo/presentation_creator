package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"presentation_creator/pkg/presentation"
)

// Storage handles file operations for presentations
type Storage struct {
	rootDir string
}

// NewStorage creates a new Storage instance
func NewStorage(rootDir string) *Storage {
	return &Storage{rootDir: rootDir}
}

// GetPresentationPath returns the directory path for a presentation
func (s *Storage) GetPresentationPath(id string) string {
	return filepath.Join(s.rootDir, id)
}

// GetSlidesDir returns the path to the slides directory
func (s *Storage) GetSlidesDir(id string) string {
	return filepath.Join(s.GetPresentationPath(id), "slides")
}

// GetSlidePath returns the path to a specific slide HTML file
func (s *Storage) GetSlidePath(id string, slideNumber int) string {
	return filepath.Join(s.GetSlidesDir(id), fmt.Sprintf("%d.html", slideNumber))
}

// GetMetadataPath returns the path to the metadata.json file
func (s *Storage) GetMetadataPath(id string) string {
	return filepath.Join(s.GetPresentationPath(id), "metadata.json")
}

// GetMediaDir returns the path to the media directory
func (s *Storage) GetMediaDir(id string) string {
	return filepath.Join(s.GetPresentationPath(id), "media")
}

// PresentationExists checks if a presentation exists
func (s *Storage) PresentationExists(id string) bool {
	metaPath := s.GetMetadataPath(id)
	_, err := os.Stat(metaPath)
	return err == nil
}

// CreatePresentation creates a new presentation on disk
func (s *Storage) CreatePresentation(p *presentation.Presentation, slides []string) error {
	presPath := s.GetPresentationPath(p.ID)
	if err := os.MkdirAll(presPath, 0755); err != nil {
		return fmt.Errorf("failed to create presentation directory: %w", err)
	}

	slidesDir := s.GetSlidesDir(p.ID)
	if err := os.MkdirAll(slidesDir, 0755); err != nil {
		return fmt.Errorf("failed to create slides directory: %w", err)
	}

	mediaDir := s.GetMediaDir(p.ID)
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return fmt.Errorf("failed to create media directory: %w", err)
	}

	for i, html := range slides {
		slidePath := s.GetSlidePath(p.ID, i+1)
		if err := os.WriteFile(slidePath, []byte(html), 0644); err != nil {
			return fmt.Errorf("failed to write slide %d: %w", i+1, err)
		}
	}

	metadata := presentation.Metadata{
		Name:       p.Name,
		Width:      presentation.SlideWidth,
		Height:     presentation.SlideHeight,
		SlideCount: len(slides),
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
	if err := s.writeMetadata(p.ID, &metadata); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// UpdateSlide updates a single slide's HTML content
func (s *Storage) UpdateSlide(id string, slideNumber int, htmlContent string) error {
	slidePath := s.GetSlidePath(id, slideNumber)
	if err := os.WriteFile(slidePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write slide %d: %w", slideNumber, err)
	}
	return nil
}

// AddSlide inserts a new slide at the given position, shifting others
func (s *Storage) AddSlide(id string, position int, htmlContent string, currentCount int) error {
	slidesDir := s.GetSlidesDir(id)

	// Shift existing slides from end to position to make room
	for i := currentCount; i >= position; i-- {
		oldPath := filepath.Join(slidesDir, fmt.Sprintf("%d.html", i))
		newPath := filepath.Join(slidesDir, fmt.Sprintf("%d.html", i+1))
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("failed to shift slide %d: %w", i, err)
		}
	}

	// Write the new slide
	slidePath := filepath.Join(slidesDir, fmt.Sprintf("%d.html", position))
	if err := os.WriteFile(slidePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write new slide at position %d: %w", position, err)
	}

	return nil
}

// DeleteSlide removes a slide and renumbers remaining slides
func (s *Storage) DeleteSlide(id string, slideNumber int, currentCount int) error {
	slidesDir := s.GetSlidesDir(id)

	// Remove the target slide
	targetPath := filepath.Join(slidesDir, fmt.Sprintf("%d.html", slideNumber))
	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("failed to delete slide %d: %w", slideNumber, err)
	}

	// Shift remaining slides down
	for i := slideNumber + 1; i <= currentCount; i++ {
		oldPath := filepath.Join(slidesDir, fmt.Sprintf("%d.html", i))
		newPath := filepath.Join(slidesDir, fmt.Sprintf("%d.html", i-1))
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("failed to renumber slide %d: %w", i, err)
		}
	}

	return nil
}

// GetSlide reads a single slide's HTML content
func (s *Storage) GetSlide(id string, slideNumber int) (string, error) {
	slidePath := s.GetSlidePath(id, slideNumber)
	data, err := os.ReadFile(slidePath)
	if err != nil {
		return "", fmt.Errorf("failed to read slide %d: %w", slideNumber, err)
	}
	return string(data), nil
}

// GetAllSlides reads all slides in order
func (s *Storage) GetAllSlides(id string, slideCount int) ([]presentation.Slide, error) {
	slides := make([]presentation.Slide, slideCount)
	for i := 1; i <= slideCount; i++ {
		content, err := s.GetSlide(id, i)
		if err != nil {
			return nil, err
		}
		slides[i-1] = presentation.Slide{
			Number:      i,
			HTMLContent: content,
		}
	}
	return slides, nil
}

// GetPresentation retrieves a presentation from disk
func (s *Storage) GetPresentation(id string) (*presentation.Presentation, error) {
	if !s.PresentationExists(id) {
		return nil, fmt.Errorf("presentation %s does not exist", id)
	}

	metadata, err := s.readMetadata(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	return &presentation.Presentation{
		ID:         id,
		Name:       metadata.Name,
		SlideCount: metadata.SlideCount,
		CreatedAt:  metadata.CreatedAt,
		UpdatedAt:  metadata.UpdatedAt,
	}, nil
}

// ListPresentations returns all presentations
func (s *Storage) ListPresentations() ([]*presentation.PresentationInfo, error) {
	entries, err := os.ReadDir(s.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read root directory: %w", err)
	}

	var presentations []*presentation.PresentationInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		if !s.PresentationExists(id) {
			continue
		}

		metadata, err := s.readMetadata(id)
		if err != nil {
			continue
		}

		presentations = append(presentations, &presentation.PresentationInfo{
			ID:         id,
			Name:       metadata.Name,
			SlideCount: metadata.SlideCount,
			CreatedAt:  metadata.CreatedAt,
			UpdatedAt:  metadata.UpdatedAt,
			FilePath:   s.GetPresentationPath(id),
		})
	}

	return presentations, nil
}

// CopyMediaFile copies a media file to the presentation's media directory
func (s *Storage) CopyMediaFile(id, sourcePath string) (string, error) {
	if !s.PresentationExists(id) {
		return "", fmt.Errorf("presentation %s does not exist", id)
	}

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	filename := filepath.Base(sourcePath)
	mediaDir := s.GetMediaDir(id)
	destPath := filepath.Join(mediaDir, filename)

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return filepath.Join("media", filename), nil
}

// UpdateMetadata updates the metadata file
func (s *Storage) UpdateMetadata(id string, metadata *presentation.Metadata) error {
	return s.writeMetadata(id, metadata)
}

// CountSlides counts slide files on disk
func (s *Storage) CountSlides(id string) (int, error) {
	slidesDir := s.GetSlidesDir(id)
	entries, err := os.ReadDir(slidesDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read slides directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			count++
		}
	}
	return count, nil
}

// GetSlideNumbers returns sorted slide numbers from disk
func (s *Storage) GetSlideNumbers(id string) ([]int, error) {
	slidesDir := s.GetSlidesDir(id)
	entries, err := os.ReadDir(slidesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read slides directory: %w", err)
	}

	var numbers []int
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasSuffix(name, ".html") {
			numStr := strings.TrimSuffix(name, ".html")
			if n, err := strconv.Atoi(numStr); err == nil {
				numbers = append(numbers, n)
			}
		}
	}

	sort.Ints(numbers)
	return numbers, nil
}

func (s *Storage) writeMetadata(id string, metadata *presentation.Metadata) error {
	metadataPath := s.GetMetadataPath(id)
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	return os.WriteFile(metadataPath, data, 0644)
}

func (s *Storage) readMetadata(id string) (*presentation.Metadata, error) {
	metadataPath := s.GetMetadataPath(id)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata presentation.Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

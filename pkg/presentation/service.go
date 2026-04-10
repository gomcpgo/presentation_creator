package presentation

import (
	"fmt"
	"time"
)

// StorageInterface defines the storage operations needed by the service
type StorageInterface interface {
	PresentationExists(id string) bool
	CreatePresentation(p *Presentation, slides []string) error
	UpdateSlide(id string, slideNumber int, htmlContent string) error
	AddSlide(id string, position int, htmlContent string, currentCount int) error
	DeleteSlide(id string, slideNumber int, currentCount int) error
	GetSlide(id string, slideNumber int) (string, error)
	GetAllSlides(id string, slideCount int) ([]Slide, error)
	GetPresentation(id string) (*Presentation, error)
	ListPresentations() ([]*PresentationInfo, error)
	CopyMediaFile(id, sourcePath string) (string, error)
	UpdateMetadata(id string, metadata *Metadata) error
	GetPresentationPath(id string) string
	GetSlidesDir(id string) string
}

// Service provides presentation operations
type Service struct {
	storage StorageInterface
}

// NewService creates a new presentation service
func NewService(storage StorageInterface) *Service {
	return &Service{storage: storage}
}

// CreatePresentation creates a new presentation with initial slides
func (s *Service) CreatePresentation(name string, slides []string) (*Presentation, error) {
	if name == "" {
		return nil, fmt.Errorf("presentation name cannot be empty")
	}
	if len(slides) == 0 {
		return nil, fmt.Errorf("at least one slide is required")
	}

	id := GenerateID(name, s.storage.PresentationExists)

	now := time.Now()
	p := &Presentation{
		ID:         id,
		Name:       name,
		SlideCount: len(slides),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.storage.CreatePresentation(p, slides); err != nil {
		return nil, fmt.Errorf("failed to create presentation: %w", err)
	}

	return p, nil
}

// UpdateSlide updates a single slide's HTML content
func (s *Service) UpdateSlide(id string, slideNumber int, htmlContent string) (*Presentation, error) {
	if !ValidateID(id) {
		return nil, fmt.Errorf("invalid presentation ID: %s", id)
	}
	if htmlContent == "" {
		return nil, fmt.Errorf("HTML content cannot be empty")
	}

	p, err := s.storage.GetPresentation(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get presentation: %w", err)
	}

	if slideNumber < 1 || slideNumber > p.SlideCount {
		return nil, fmt.Errorf("slide number %d out of range (1-%d)", slideNumber, p.SlideCount)
	}

	if err := s.storage.UpdateSlide(id, slideNumber, htmlContent); err != nil {
		return nil, fmt.Errorf("failed to update slide: %w", err)
	}

	p.UpdatedAt = time.Now()
	s.updateMetadata(p)

	return p, nil
}

// AddSlide adds a new slide at the given position
func (s *Service) AddSlide(id string, htmlContent string, position int) (*Presentation, error) {
	if !ValidateID(id) {
		return nil, fmt.Errorf("invalid presentation ID: %s", id)
	}
	if htmlContent == "" {
		return nil, fmt.Errorf("HTML content cannot be empty")
	}

	p, err := s.storage.GetPresentation(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get presentation: %w", err)
	}

	// Default to end if position is 0 or out of range
	if position <= 0 || position > p.SlideCount+1 {
		position = p.SlideCount + 1
	}

	if err := s.storage.AddSlide(id, position, htmlContent, p.SlideCount); err != nil {
		return nil, fmt.Errorf("failed to add slide: %w", err)
	}

	p.SlideCount++
	p.UpdatedAt = time.Now()
	s.updateMetadata(p)

	return p, nil
}

// DeleteSlide removes a slide and renumbers remaining slides
func (s *Service) DeleteSlide(id string, slideNumber int) (*Presentation, error) {
	if !ValidateID(id) {
		return nil, fmt.Errorf("invalid presentation ID: %s", id)
	}

	p, err := s.storage.GetPresentation(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get presentation: %w", err)
	}

	if p.SlideCount <= 1 {
		return nil, fmt.Errorf("cannot delete the last slide")
	}

	if slideNumber < 1 || slideNumber > p.SlideCount {
		return nil, fmt.Errorf("slide number %d out of range (1-%d)", slideNumber, p.SlideCount)
	}

	if err := s.storage.DeleteSlide(id, slideNumber, p.SlideCount); err != nil {
		return nil, fmt.Errorf("failed to delete slide: %w", err)
	}

	p.SlideCount--
	p.UpdatedAt = time.Now()
	s.updateMetadata(p)

	return p, nil
}

// GetPresentation retrieves a presentation with all slide contents
func (s *Service) GetPresentation(id string) (*Presentation, []Slide, error) {
	if !ValidateID(id) {
		return nil, nil, fmt.Errorf("invalid presentation ID: %s", id)
	}

	p, err := s.storage.GetPresentation(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get presentation: %w", err)
	}

	slides, err := s.storage.GetAllSlides(id, p.SlideCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get slides: %w", err)
	}

	return p, slides, nil
}

// ListPresentations returns all presentations
func (s *Service) ListPresentations() ([]*PresentationInfo, error) {
	return s.storage.ListPresentations()
}

// AddMedia adds a media file to a presentation
func (s *Service) AddMedia(id, sourcePath string) (string, error) {
	if !ValidateID(id) {
		return "", fmt.Errorf("invalid presentation ID: %s", id)
	}
	if sourcePath == "" {
		return "", fmt.Errorf("source path cannot be empty")
	}

	return s.storage.CopyMediaFile(id, sourcePath)
}

// GetPresentationPath returns the absolute path to the presentation directory
func (s *Service) GetPresentationPath(id string) string {
	return s.storage.GetPresentationPath(id)
}

// GetSlidesDir returns the absolute path to the slides directory
func (s *Service) GetSlidesDir(id string) string {
	return s.storage.GetSlidesDir(id)
}

func (s *Service) updateMetadata(p *Presentation) {
	metadata := &Metadata{
		Name:       p.Name,
		Width:      SlideWidth,
		Height:     SlideHeight,
		SlideCount: p.SlideCount,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
	// Best-effort metadata update
	s.storage.UpdateMetadata(p.ID, metadata)
}

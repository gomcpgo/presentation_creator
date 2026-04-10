package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"presentation_creator/pkg/config"
	"presentation_creator/pkg/presentation"
	"presentation_creator/pkg/storage"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// ScreenshotService defines the interface for screenshot functionality
type ScreenshotService interface {
	TakeScreenshot(htmlDir string, htmlFile string, width, height int, outputPath string) error
}

// ExportService defines the interface for export functionality
type ExportService interface {
	ExportPPTX(slideImages []string, outputPath string) error
	ExportPDF(slideImages []string, outputPath string) error
}

// Handler implements the MCP protocol for Presentation Creator
type Handler struct {
	config        *config.Config
	presSvc       *presentation.Service
	screenshotSvc ScreenshotService
	exportSvc     ExportService
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config, screenshotSvc ScreenshotService, exportSvc ExportService) *Handler {
	store := storage.NewStorage(cfg.RootDir)
	presSvc := presentation.NewService(store)

	return &Handler{
		config:        cfg,
		presSvc:       presSvc,
		screenshotSvc: screenshotSvc,
		exportSvc:     exportSvc,
	}
}

// ListTools returns the list of available tools
func (h *Handler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	return &protocol.ListToolsResponse{Tools: h.GetTools()}, nil
}

// CallTool handles tool invocations
func (h *Handler) CallTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResponse, error) {
	switch req.Name {
	case "create_presentation":
		return h.handleCreatePresentation(ctx, req.Arguments)
	case "update_slide":
		return h.handleUpdateSlide(ctx, req.Arguments)
	case "add_slide":
		return h.handleAddSlide(ctx, req.Arguments)
	case "delete_slide":
		return h.handleDeleteSlide(ctx, req.Arguments)
	case "get_presentation":
		return h.handleGetPresentation(ctx, req.Arguments)
	case "list_presentations":
		return h.handleListPresentations(ctx, req.Arguments)
	case "add_media":
		return h.handleAddMedia(ctx, req.Arguments)
	case "export_presentation":
		return h.handleExportPresentation(ctx, req.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}
}

func (h *Handler) handleCreatePresentation(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required and must be a string")
	}

	slidesRaw, ok := args["slides"].([]interface{})
	if !ok || len(slidesRaw) == 0 {
		return nil, fmt.Errorf("slides is required and must be a non-empty array")
	}

	slides := make([]string, len(slidesRaw))
	for i, s := range slidesRaw {
		html, ok := s.(string)
		if !ok || html == "" {
			return nil, fmt.Errorf("slide %d must be a non-empty string", i+1)
		}
		slides[i] = html
	}

	p, err := h.presSvc.CreatePresentation(name, slides)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to create presentation: %v", err)), nil
	}

	// Process optional media_files
	var mediaPaths map[string]string
	if mediaFilesRaw, ok := args["media_files"].([]interface{}); ok && len(mediaFilesRaw) > 0 {
		mediaPaths = make(map[string]string)
		for _, mf := range mediaFilesRaw {
			sourcePath, ok := mf.(string)
			if !ok || sourcePath == "" {
				continue
			}
			relativePath, err := h.presSvc.AddMedia(p.ID, sourcePath)
			if err != nil {
				return h.errorResponse(fmt.Sprintf("Failed to add media file %s: %v", sourcePath, err)), nil
			}
			mediaPaths[sourcePath] = relativePath
		}
	}

	result := map[string]interface{}{
		"status":          "succeeded",
		"presentation_id": p.ID,
		"name":            p.Name,
		"width":           presentation.SlideWidth,
		"height":          presentation.SlideHeight,
		"slide_count":     p.SlideCount,
		"file_path":       h.presSvc.GetPresentationPath(p.ID),
		"created_at":      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if len(mediaPaths) > 0 {
		result["media_paths"] = mediaPaths
	}

	return h.successResponse(result), nil
}

func (h *Handler) handleUpdateSlide(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	presID, ok := args["presentation_id"].(string)
	if !ok || presID == "" {
		return nil, fmt.Errorf("presentation_id is required")
	}

	slideNumFloat, ok := args["slide_number"].(float64)
	if !ok {
		return nil, fmt.Errorf("slide_number is required and must be an integer")
	}
	slideNumber := int(slideNumFloat)

	htmlContent, ok := args["html_content"].(string)
	if !ok || htmlContent == "" {
		return nil, fmt.Errorf("html_content is required")
	}

	p, err := h.presSvc.UpdateSlide(presID, slideNumber, htmlContent)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to update slide: %v", err)), nil
	}

	result := map[string]interface{}{
		"status":          "succeeded",
		"presentation_id": p.ID,
		"name":            p.Name,
		"slide_number":    slideNumber,
		"slide_count":     p.SlideCount,
		"file_path":       h.presSvc.GetPresentationPath(p.ID),
		"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return h.successResponse(result), nil
}

func (h *Handler) handleAddSlide(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	presID, ok := args["presentation_id"].(string)
	if !ok || presID == "" {
		return nil, fmt.Errorf("presentation_id is required")
	}

	htmlContent, ok := args["html_content"].(string)
	if !ok || htmlContent == "" {
		return nil, fmt.Errorf("html_content is required")
	}

	position := 0
	if posFloat, ok := args["position"].(float64); ok {
		position = int(posFloat)
	}

	p, err := h.presSvc.AddSlide(presID, htmlContent, position)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to add slide: %v", err)), nil
	}

	result := map[string]interface{}{
		"status":          "succeeded",
		"presentation_id": p.ID,
		"name":            p.Name,
		"slide_count":     p.SlideCount,
		"file_path":       h.presSvc.GetPresentationPath(p.ID),
		"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return h.successResponse(result), nil
}

func (h *Handler) handleDeleteSlide(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	presID, ok := args["presentation_id"].(string)
	if !ok || presID == "" {
		return nil, fmt.Errorf("presentation_id is required")
	}

	slideNumFloat, ok := args["slide_number"].(float64)
	if !ok {
		return nil, fmt.Errorf("slide_number is required and must be an integer")
	}
	slideNumber := int(slideNumFloat)

	p, err := h.presSvc.DeleteSlide(presID, slideNumber)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete slide: %v", err)), nil
	}

	result := map[string]interface{}{
		"status":          "succeeded",
		"presentation_id": p.ID,
		"name":            p.Name,
		"slide_count":     p.SlideCount,
		"file_path":       h.presSvc.GetPresentationPath(p.ID),
		"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return h.successResponse(result), nil
}

func (h *Handler) handleGetPresentation(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	presID, ok := args["presentation_id"].(string)
	if !ok || presID == "" {
		return nil, fmt.Errorf("presentation_id is required")
	}

	p, slides, err := h.presSvc.GetPresentation(presID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get presentation: %v", err)), nil
	}

	slidesData := make([]map[string]interface{}, len(slides))
	for i, s := range slides {
		slidesData[i] = map[string]interface{}{
			"number":       s.Number,
			"html_content": s.HTMLContent,
		}
	}

	result := map[string]interface{}{
		"status":          "succeeded",
		"presentation_id": p.ID,
		"name":            p.Name,
		"width":           presentation.SlideWidth,
		"height":          presentation.SlideHeight,
		"slide_count":     p.SlideCount,
		"slides":          slidesData,
		"file_path":       h.presSvc.GetPresentationPath(p.ID),
		"created_at":      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return h.successResponse(result), nil
}

func (h *Handler) handleListPresentations(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	presentations, err := h.presSvc.ListPresentations()
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to list presentations: %v", err)), nil
	}

	presList := make([]map[string]interface{}, len(presentations))
	for i, p := range presentations {
		presList[i] = map[string]interface{}{
			"presentation_id": p.ID,
			"name":            p.Name,
			"slide_count":     p.SlideCount,
			"file_path":       p.FilePath,
			"created_at":      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at":      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	result := map[string]interface{}{
		"status":        "succeeded",
		"count":         len(presList),
		"presentations": presList,
	}

	return h.successResponse(result), nil
}

func (h *Handler) handleAddMedia(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	presID, ok := args["presentation_id"].(string)
	if !ok || presID == "" {
		return nil, fmt.Errorf("presentation_id is required")
	}

	sourcePath, ok := args["source_path"].(string)
	if !ok || sourcePath == "" {
		return nil, fmt.Errorf("source_path is required")
	}

	relativePath, err := h.presSvc.AddMedia(presID, sourcePath)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to add media: %v", err)), nil
	}

	result := map[string]interface{}{
		"status":          "succeeded",
		"presentation_id": presID,
		"relative_path":   relativePath,
	}

	return h.successResponse(result), nil
}

func (h *Handler) handleExportPresentation(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	presID, ok := args["presentation_id"].(string)
	if !ok || presID == "" {
		return nil, fmt.Errorf("presentation_id is required")
	}

	format, ok := args["format"].(string)
	if !ok || (format != "pptx" && format != "pdf") {
		return nil, fmt.Errorf("format must be 'pptx' or 'pdf'")
	}

	outputPath, ok := args["output_path"].(string)
	if !ok || outputPath == "" {
		return nil, fmt.Errorf("output_path is required")
	}

	p, _, err := h.presSvc.GetPresentation(presID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get presentation: %v", err)), nil
	}

	// Screenshot each slide
	presPath := h.presSvc.GetPresentationPath(presID)
	slidesDir := h.presSvc.GetSlidesDir(presID)
	slideImages := make([]string, p.SlideCount)

	for i := 1; i <= p.SlideCount; i++ {
		imgPath := fmt.Sprintf("%s/slide_%d.png", presPath, i)
		slideFile := fmt.Sprintf("%d.html", i)
		if err := h.screenshotSvc.TakeScreenshot(slidesDir, slideFile, presentation.SlideWidth, presentation.SlideHeight, imgPath); err != nil {
			return h.errorResponse(fmt.Sprintf("Failed to screenshot slide %d: %v", i, err)), nil
		}
		slideImages[i-1] = imgPath
	}

	// Assemble into output format
	var exportErr error
	switch format {
	case "pptx":
		exportErr = h.exportSvc.ExportPPTX(slideImages, outputPath)
	case "pdf":
		exportErr = h.exportSvc.ExportPDF(slideImages, outputPath)
	}

	// Clean up temporary slide images
	for _, img := range slideImages {
		os.Remove(img)
	}

	if exportErr != nil {
		return h.errorResponse(fmt.Sprintf("Failed to export %s: %v", format, exportErr)), nil
	}

	result := map[string]interface{}{
		"status":          "succeeded",
		"presentation_id": presID,
		"format":          format,
		"output_path":     outputPath,
	}

	return h.successResponse(result), nil
}

func (h *Handler) successResponse(data map[string]interface{}) *protocol.CallToolResponse {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{Type: "text", Text: string(jsonData)},
		},
	}
}

func (h *Handler) errorResponse(errorMsg string) *protocol.CallToolResponse {
	data := map[string]interface{}{
		"status": "failed",
		"error":  errorMsg,
	}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{Type: "text", Text: string(jsonData)},
		},
	}
}

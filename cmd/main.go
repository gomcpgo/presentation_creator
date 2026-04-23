package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"presentation_creator/pkg/config"
	mcpHandler "presentation_creator/pkg/handler"
	"presentation_creator/pkg/pptx"
	"presentation_creator/pkg/screenshot"

	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/mcp/pkg/server"
)

//go:embed icon.svg
var iconSVG []byte

func main() {
	var (
		createPres   string
		updateSlide  string
		addSlide     string
		deleteSlide  string
		getPres      string
		listPres     bool
		exportPres   string
		exportFormat string
		exportOutput string
		addMedia     string
		mediaPath    string
		slideNum     int
		htmlContent  string
		slidesJSON   string
		position     int
	)

	flag.StringVar(&createPres, "create", "", "Create a new presentation with the specified name")
	flag.StringVar(&slidesJSON, "slides", "", "JSON array of HTML strings for slides (used with -create)")
	flag.StringVar(&updateSlide, "update", "", "Update slide in presentation (specify presentation ID)")
	flag.StringVar(&addSlide, "add-slide", "", "Add slide to presentation (specify presentation ID)")
	flag.StringVar(&deleteSlide, "delete-slide", "", "Delete slide from presentation (specify presentation ID)")
	flag.IntVar(&slideNum, "slide", 0, "Slide number (used with -update, -delete-slide)")
	flag.IntVar(&position, "position", 0, "Position for new slide (used with -add-slide)")
	flag.StringVar(&htmlContent, "html", "", "HTML content for slide operations")
	flag.StringVar(&getPres, "get", "", "Get presentation by ID")
	flag.BoolVar(&listPres, "list", false, "List all presentations")
	flag.StringVar(&exportPres, "export", "", "Export presentation by ID")
	flag.StringVar(&exportFormat, "format", "pptx", "Export format: pptx or pdf")
	flag.StringVar(&exportOutput, "output", "", "Output path for export")
	flag.StringVar(&addMedia, "add-media", "", "Add media to presentation (specify presentation ID)")
	flag.StringVar(&mediaPath, "media-path", "", "Path to media file")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	screenshotSvc := screenshot.NewScreenshotter()
	exportSvc := pptx.NewBuilder()
	h := mcpHandler.NewHandler(cfg, screenshotSvc, exportSvc)
	ctx := context.Background()

	// Terminal mode operations
	if createPres != "" {
		if slidesJSON == "" {
			log.Fatal("--slides is required when creating a presentation (JSON array of HTML strings)")
		}
		var slides []interface{}
		if err := json.Unmarshal([]byte(slidesJSON), &slides); err != nil {
			log.Fatalf("Failed to parse --slides JSON: %v", err)
		}
		runTerminalCommand(ctx, h, "create_presentation", map[string]interface{}{
			"name":   createPres,
			"slides": slides,
		})
		return
	}

	if updateSlide != "" {
		if htmlContent == "" || slideNum <= 0 {
			log.Fatal("--html and --slide are required when updating a slide")
		}
		runTerminalCommand(ctx, h, "update_slide", map[string]interface{}{
			"presentation_id": updateSlide,
			"slide_number":    float64(slideNum),
			"html_content":    htmlContent,
		})
		return
	}

	if addSlide != "" {
		if htmlContent == "" {
			log.Fatal("--html is required when adding a slide")
		}
		args := map[string]interface{}{
			"presentation_id": addSlide,
			"html_content":    htmlContent,
		}
		if position > 0 {
			args["position"] = float64(position)
		}
		runTerminalCommand(ctx, h, "add_slide", args)
		return
	}

	if deleteSlide != "" {
		if slideNum <= 0 {
			log.Fatal("--slide is required when deleting a slide")
		}
		runTerminalCommand(ctx, h, "delete_slide", map[string]interface{}{
			"presentation_id": deleteSlide,
			"slide_number":    float64(slideNum),
		})
		return
	}

	if getPres != "" {
		runTerminalCommand(ctx, h, "get_presentation", map[string]interface{}{
			"presentation_id": getPres,
		})
		return
	}

	if listPres {
		runTerminalCommand(ctx, h, "list_presentations", map[string]interface{}{})
		return
	}

	if exportPres != "" {
		if exportOutput == "" {
			log.Fatal("--output is required when exporting")
		}
		runTerminalCommand(ctx, h, "export_presentation", map[string]interface{}{
			"presentation_id": exportPres,
			"format":          exportFormat,
			"output_path":     exportOutput,
		})
		return
	}

	if addMedia != "" {
		if mediaPath == "" {
			log.Fatal("--media-path is required when adding media")
		}
		runTerminalCommand(ctx, h, "add_media", map[string]interface{}{
			"presentation_id": addMedia,
			"source_path":     mediaPath,
		})
		return
	}

	// MCP Server mode (default)
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(h)

	srv := server.New(server.Options{
		Name:     "presentation-creator",
		Title:    "Presentation Creator",
		Version:  "1.0.0",
		Icons:    protocol.IconFromSVG(iconSVG),
		Registry: registry,
	})

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runTerminalCommand(ctx context.Context, h *mcpHandler.Handler, toolName string, args map[string]interface{}) {
	req := &protocol.CallToolRequest{
		Name:      toolName,
		Arguments: args,
	}

	resp, err := h.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	for _, content := range resp.Content {
		if content.Type == "text" {
			var data interface{}
			if err := json.Unmarshal([]byte(content.Text), &data); err == nil {
				pretty, _ := json.MarshalIndent(data, "", "  ")
				fmt.Println(string(pretty))
			} else {
				fmt.Println(content.Text)
			}
		}
	}
}

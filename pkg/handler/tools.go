package handler

import (
	"encoding/json"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// GetTools returns the list of available MCP tools
func (h *Handler) GetTools() []protocol.Tool {
	return []protocol.Tool{
		{
			Name:        "create_presentation",
			Description: "Create a new HTML slide presentation with fixed 1920x1080 (16:9) dimensions. Each slide is a full HTML/CSS page rendered at these dimensions. The LLM should generate visually rich slides using HTML/CSS — backgrounds, gradients, images, charts, custom fonts via Google Fonts <link> tags. To include images, pass their absolute file paths in media_files. Reference them in HTML as media/filename.ext (e.g., <img src=\"media/chart.png\"> or background-image: url('media/logo.svg')).",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "Name of the presentation"
					},
					"slides": {
						"type": "array",
						"items": { "type": "string" },
						"description": "Array of HTML content strings, one per slide. Each slide is a full HTML page rendered at 1920x1080."
					},
					"media_files": {
						"type": "array",
						"items": { "type": "string" },
						"description": "Optional list of absolute file paths to copy into the presentation's media folder. Each file becomes available as media/filename.ext in slide HTML."
					}
				},
				"required": ["name", "slides"]
			}`),
		},
		{
			Name:        "update_slide",
			Description: "Update the HTML content of a single slide in a presentation.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"presentation_id": {
						"type": "string",
						"description": "The unique presentation ID"
					},
					"slide_number": {
						"type": "integer",
						"description": "1-based slide number to update"
					},
					"html_content": {
						"type": "string",
						"description": "The new HTML/CSS content for this slide"
					}
				},
				"required": ["presentation_id", "slide_number", "html_content"]
			}`),
		},
		{
			Name:        "add_slide",
			Description: "Add a new slide to a presentation at a specified position.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"presentation_id": {
						"type": "string",
						"description": "The unique presentation ID"
					},
					"html_content": {
						"type": "string",
						"description": "HTML/CSS content for the new slide"
					},
					"position": {
						"type": "integer",
						"description": "Position to insert the slide (1-based). Defaults to end if omitted."
					}
				},
				"required": ["presentation_id", "html_content"]
			}`),
		},
		{
			Name:        "delete_slide",
			Description: "Remove a slide from a presentation. Remaining slides are renumbered.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"presentation_id": {
						"type": "string",
						"description": "The unique presentation ID"
					},
					"slide_number": {
						"type": "integer",
						"description": "1-based slide number to delete"
					}
				},
				"required": ["presentation_id", "slide_number"]
			}`),
		},
		{
			Name:        "get_presentation",
			Description: "Retrieve a presentation's metadata and all slide contents.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"presentation_id": {
						"type": "string",
						"description": "The unique presentation ID"
					}
				},
				"required": ["presentation_id"]
			}`),
		},
		{
			Name:        "list_presentations",
			Description: "List all presentations with their metadata.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		{
			Name:        "add_media",
			Description: "Add a media file (image, SVG, etc.) to a presentation's media folder. Returns the relative path to use in slide HTML (e.g., media/photo.jpg).",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"presentation_id": {
						"type": "string",
						"description": "The unique presentation ID"
					},
					"source_path": {
						"type": "string",
						"description": "Absolute path to the source media file"
					}
				},
				"required": ["presentation_id", "source_path"]
			}`),
		},
		{
			Name:        "export_presentation",
			Description: "Export a presentation to PPTX or PDF. Each slide is rendered as a high-quality image via headless Chrome and assembled into the output format.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"presentation_id": {
						"type": "string",
						"description": "The unique presentation ID"
					},
					"format": {
						"type": "string",
						"enum": ["pptx", "pdf"],
						"description": "Export format: pptx or pdf"
					},
					"output_path": {
						"type": "string",
						"description": "Absolute path for the output file"
					}
				},
				"required": ["presentation_id", "format", "output_path"]
			}`),
		},
	}
}

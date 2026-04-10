# Presentation Creator MCP Server

An MCP server that enables LLMs to create visually rich slide presentations using HTML/CSS. Each slide is a full HTML page rendered at 1920x1080 (16:9). Presentations are exported as PPTX or PDF by screenshotting each slide via headless Chrome.

## Features

- Create multi-slide presentations with HTML/CSS
- Update, add, or delete individual slides
- Add images and media files for use in slides
- Export to PPTX (PowerPoint) or PDF
- Terminal mode for testing
- Fixed 1920x1080 slide dimensions (standard 16:9 widescreen)

## How It Works

1. The LLM generates HTML/CSS for each slide with full visual control (backgrounds, gradients, images, custom fonts)
2. Each slide is stored as a separate HTML file
3. On export, each slide is screenshotted at 1920x1080 via headless Chrome (2x resolution)
4. Screenshots are assembled into PPTX (using XML/ZIP) or PDF

## Presentation Structure

```
{ROOT_DIR}/quarterly-review-a3f9/
├── metadata.json        # Presentation metadata
├── slides/
│   ├── 1.html           # Slide 1 HTML
│   ├── 2.html           # Slide 2 HTML
│   └── 3.html           # Slide 3 HTML
├── media/               # Shared media files
│   ├── chart.png
│   └── logo.svg
└── exports/             # Generated exports
```

## Configuration

Set the root directory via environment variable:
```bash
export PRESENTATION_CREATOR_ROOT_DIR="/path/to/presentations"
```

Default: `~/.savant_presentations`

## Building

```bash
./run.sh install  # Install dependencies
./run.sh build    # Build binary to bin/presentation_creator
```

## Terminal Mode (Testing)

```bash
# Create a presentation
./run.sh create "Q4 Review" '["<h1>Q4 Review</h1>","<h1>Revenue</h1>"]'

# List presentations
./run.sh list

# Get presentation
./run.sh get q4-review-a3f9

# Update a slide
./run.sh update-slide q4-review-a3f9 1 "<h1>Updated Title</h1>"

# Add a slide
./run.sh add-slide q4-review-a3f9 "<h1>New Slide</h1>"

# Delete a slide
./run.sh delete-slide q4-review-a3f9 3

# Add media
./run.sh add-media q4-review-a3f9 /path/to/chart.png

# Export to PPTX
./run.sh export q4-review-a3f9 ~/Desktop/q4-review.pptx

# Export to PDF
./run.sh export q4-review-a3f9 ~/Desktop/q4-review.pdf pdf
```

## MCP Tools

### create_presentation
Create a new presentation with initial slides.

**Parameters:**
- `name` (string, required): Presentation name
- `slides` (array of strings, required): HTML content for each slide
- `media_files` (array of strings, optional): Absolute paths to media files

**Returns:**
```json
{
  "status": "succeeded",
  "presentation_id": "q4-review-a3f9",
  "name": "Q4 Review",
  "width": 1920,
  "height": 1080,
  "slide_count": 3,
  "file_path": "/path/to/q4-review-a3f9"
}
```

### update_slide
Update a single slide's HTML content.

**Parameters:**
- `presentation_id` (string, required): Presentation ID
- `slide_number` (integer, required): 1-based slide number
- `html_content` (string, required): New HTML content

### add_slide
Add a new slide at a specified position.

**Parameters:**
- `presentation_id` (string, required): Presentation ID
- `html_content` (string, required): HTML content
- `position` (integer, optional): Insert position (defaults to end)

### delete_slide
Remove a slide. Remaining slides are renumbered.

**Parameters:**
- `presentation_id` (string, required): Presentation ID
- `slide_number` (integer, required): Slide number to delete

### get_presentation
Retrieve a presentation with all slide contents.

**Parameters:**
- `presentation_id` (string, required): Presentation ID

### list_presentations
List all presentations with metadata.

### add_media
Add a media file to a presentation's media folder.

**Parameters:**
- `presentation_id` (string, required): Presentation ID
- `source_path` (string, required): Absolute path to media file

**Returns:** `relative_path` (e.g., `media/chart.png`) for use in HTML.

### export_presentation
Export to PPTX or PDF.

**Parameters:**
- `presentation_id` (string, required): Presentation ID
- `format` (string, required): `"pptx"` or `"pdf"`
- `output_path` (string, required): Output file path

## Export Requirements

Headless Chrome is required for export. The server uses Rod (Go Chrome DevTools wrapper) which will auto-download Chromium if a system Chrome is not found.

## Testing

```bash
./run.sh test
```

## Architecture

- `cmd/main.go` - Entry point with terminal mode
- `pkg/config/` - Configuration from env vars
- `pkg/presentation/` - Core presentation logic and types
- `pkg/storage/` - File operations (per-slide HTML, metadata)
- `pkg/screenshot/` - Headless Chrome screenshotting
- `pkg/pptx/` - PPTX and PDF assembly
- `pkg/handler/` - MCP protocol implementation

## License

MIT License

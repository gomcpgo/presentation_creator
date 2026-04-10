package pptx

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// buildPDF creates a PDF from slide images using headless Chrome.
// It builds a temporary HTML page with all slide images as full-page elements,
// then uses Chrome's PDF printer for reliable output.
func buildPDF(slideImages []string, outputPath string) error {
	if len(slideImages) == 0 {
		return fmt.Errorf("no slide images provided")
	}

	// Build HTML with all slides as pages
	var html strings.Builder
	html.WriteString(`<!DOCTYPE html><html><head><style>
@page { size: 1920px 1080px; margin: 0; }
body { margin: 0; padding: 0; }
.slide { width: 1920px; height: 1080px; page-break-after: always; overflow: hidden; }
.slide:last-child { page-break-after: auto; }
.slide img { width: 100%; height: 100%; display: block; }
</style></head><body>
`)

	for _, imgPath := range slideImages {
		imgData, err := os.ReadFile(imgPath)
		if err != nil {
			return fmt.Errorf("failed to read image %s: %w", imgPath, err)
		}
		b64 := base64.StdEncoding.EncodeToString(imgData)
		html.WriteString(fmt.Sprintf(`<div class="slide"><img src="data:image/png;base64,%s"/></div>
`, b64))
	}

	html.WriteString("</body></html>")

	// Write temp HTML
	tmpDir := filepath.Dir(outputPath)
	tmpHTML := filepath.Join(tmpDir, "temp_pdf_slides.html")
	if err := os.WriteFile(tmpHTML, []byte(html.String()), 0644); err != nil {
		return fmt.Errorf("failed to write temp HTML: %w", err)
	}
	defer os.Remove(tmpHTML)

	// Serve the HTML via local HTTP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start temp server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	fileServer := http.FileServer(http.Dir(tmpDir))
	httpServer := &http.Server{Handler: fileServer}
	go httpServer.Serve(listener)
	defer httpServer.Close()

	// Launch headless Chrome
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chromePath, _ := launcher.LookPath()
	var controlURL string
	if chromePath != "" {
		controlURL = launcher.New().Bin(chromePath).Headless(true).MustLaunch()
	} else {
		controlURL = launcher.New().Headless(true).MustLaunch()
	}

	browser := rod.New().ControlURL(controlURL).Context(ctx)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("chrome not available: %w", err)
	}
	defer browser.MustClose()

	pageURL := fmt.Sprintf("http://127.0.0.1:%d/temp_pdf_slides.html", port)
	page, err := browser.Page(proto.TargetCreateTarget{URL: pageURL})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to load page: %w", err)
	}

	// Print to PDF with landscape orientation matching slide dimensions
	pdfData, err := page.PDF(&proto.PagePrintToPDF{
		PrintBackground:   true,
		PreferCSSPageSize: true,
		MarginTop:         float64Ptr(0),
		MarginBottom:      float64Ptr(0),
		MarginLeft:        float64Ptr(0),
		MarginRight:       float64Ptr(0),
	})
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	pdfBytes, err := io.ReadAll(pdfData)
	if err != nil {
		return fmt.Errorf("failed to read PDF data: %w", err)
	}

	if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}

func float64Ptr(v float64) *float64 {
	return &v
}

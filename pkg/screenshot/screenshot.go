package screenshot

import (
	"context"
	"fmt"
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

// Screenshotter handles taking screenshots of HTML slides via headless Chrome
type Screenshotter struct {
	chromeTimeout time.Duration
}

// NewScreenshotter creates a new Screenshotter instance
func NewScreenshotter() *Screenshotter {
	return &Screenshotter{
		chromeTimeout: 30 * time.Second,
	}
}

// TakeScreenshot renders an HTML file at exact dimensions and saves as PNG.
// htmlDir is the directory to serve (for resolving relative media paths).
// htmlFile is the filename within htmlDir to screenshot.
func (s *Screenshotter) TakeScreenshot(htmlDir string, htmlFile string, width, height int, outputPath string) error {
	htmlPath := filepath.Join(htmlDir, htmlFile)
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}

	htmlContent := injectCSSReset(string(htmlBytes))

	// We need to serve the presentation root (parent of slides/) so media/ is accessible
	// The htmlDir might be the slides/ directory, so serve its parent
	serveDir := htmlDir
	if filepath.Base(htmlDir) == "slides" {
		serveDir = filepath.Dir(htmlDir)
	}

	// Write temp HTML to serveDir so relative paths (media/photo.png) resolve from the root
	tmpHTMLPath := filepath.Join(serveDir, "temp_screenshot.html")
	if err := os.WriteFile(tmpHTMLPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write temp HTML file: %w", err)
	}
	defer os.Remove(tmpHTMLPath)

	// Start a temporary local HTTP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start temp server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	fileServer := http.FileServer(http.Dir(serveDir))
	httpServer := &http.Server{Handler: fileServer}
	go httpServer.Serve(listener)
	defer httpServer.Close()

	// Launch headless Chrome
	ctx, cancel := context.WithTimeout(context.Background(), s.chromeTimeout)
	defer cancel()

	chromePath, _ := launcher.LookPath()

	var controlURL string
	if chromePath != "" {
		l := launcher.New().Bin(chromePath).Headless(true)
		controlURL = l.MustLaunch()
	} else {
		l := launcher.New().Headless(true)
		controlURL = l.MustLaunch()
	}

	browser := rod.New().ControlURL(controlURL).Context(ctx)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("chrome not available: %w", err)
	}
	defer browser.MustClose()

	pageURL := fmt.Sprintf("http://127.0.0.1:%d/temp_screenshot.html", port)
	page, err := browser.Page(proto.TargetCreateTarget{URL: pageURL})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	err = page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             width,
		Height:            height,
		DeviceScaleFactor: 2,
	})
	if err != nil {
		return fmt.Errorf("failed to set viewport: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to load page: %w", err)
	}

	// Wait for fonts to load
	_, err = page.Eval(`() => document.fonts.ready`)
	if err != nil {
		fmt.Printf("Warning: fonts.ready check failed: %v\n", err)
	}

	screenshotData, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
		Clip: &proto.PageViewport{
			X:      0,
			Y:      0,
			Width:  float64(width),
			Height: float64(height),
			Scale:  1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, screenshotData, 0644); err != nil {
		return fmt.Errorf("failed to write screenshot: %w", err)
	}

	return nil
}

// injectCSSReset injects a CSS reset for accurate viewport rendering
func injectCSSReset(htmlContent string) string {
	cssReset := `<style>html,body{margin:0;padding:0;overflow:hidden;}</style>`

	if idx := strings.Index(strings.ToLower(htmlContent), "</head>"); idx != -1 {
		return htmlContent[:idx] + cssReset + "\n" + htmlContent[idx:]
	}

	if idx := strings.Index(strings.ToLower(htmlContent), "<body"); idx != -1 {
		if endIdx := strings.Index(htmlContent[idx:], ">"); endIdx != -1 {
			insertPos := idx + endIdx + 1
			return htmlContent[:insertPos] + "\n" + cssReset + "\n" + htmlContent[insertPos:]
		}
	}

	return cssReset + "\n" + htmlContent
}

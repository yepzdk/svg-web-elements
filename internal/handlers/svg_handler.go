package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/svg-web-elements/internal/svg"
)

// SVGHandler handles requests for SVG files
type SVGHandler struct {
	processor *svg.Processor
}

// NewSVGHandler creates a new SVG handler
func NewSVGHandler(svgBasePath string) *SVGHandler {
	return &SVGHandler{
		processor: svg.NewProcessor(svgBasePath),
	}
}

// ServeHTTP handles HTTP requests for SVGs
func (h *SVGHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract the SVG name from the URL path
	path := r.URL.Path
	svgName := filepath.Base(path)
	
	log.Printf("SVG request: %s, User-Agent: %s", svgName, r.UserAgent())

	// Parse query parameters
	params, err := parseQueryParams(r.URL.Query())
	if err != nil {
		log.Printf("Error parsing parameters for %s: %v", svgName, err)
		http.Error(w, fmt.Sprintf("Invalid parameters: %v", err), http.StatusBadRequest)
		return
	}

	// Process the SVG
	svgData, err := h.processor.ProcessSVG(svgName, params)
	if err != nil {
		log.Printf("Error processing SVG %s: %v", svgName, err)
		http.Error(w, fmt.Sprintf("Error processing SVG: %v", err), http.StatusInternalServerError)
		return
	}
	
	log.Printf("Successfully processed SVG: %s, size: %d bytes", svgName, len(svgData))

	// Set content type and other headers
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	
	// Set appropriate content length
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(svgData)))
	
	// Write the SVG data
	bytesWritten, err := w.Write(svgData)
	if err != nil {
		log.Printf("Error writing SVG response for %s: %v", svgName, err)
	} else if bytesWritten != len(svgData) {
		log.Printf("Warning: Incomplete SVG write for %s: %d of %d bytes", svgName, bytesWritten, len(svgData))
	}
}

// parseQueryParams transforms URL query parameters into SVG parameters
func parseQueryParams(query url.Values) (svg.SVGParams, error) {
	params := svg.SVGParams{
		TextReplacements:  make(map[string]string),
		ColorReplacements: make(map[string]string),
	}

	// Handle width and height
	if width := query.Get("width"); width != "" {
		params.Width = width
	}
	if height := query.Get("height"); height != "" {
		params.Height = height
	}

	// Handle text replacements (format: text.element-id=value)
	for key, values := range query {
		if strings.HasPrefix(key, "text.") && len(values) > 0 {
			elementID := strings.TrimPrefix(key, "text.")
			params.TextReplacements[elementID] = values[0]
		}
		if strings.HasPrefix(key, "color.") && len(values) > 0 {
			elementID := strings.TrimPrefix(key, "color.")
			params.ColorReplacements[elementID] = values[0]
		}
	}

	// Handle external URL parameter
	if externalURL := query.Get("url"); externalURL != "" {
		// Make sure the URL is properly sanitized to prevent XSS
		sanitizedURL := strings.Replace(externalURL, "<", "&lt;", -1)
		sanitizedURL = strings.Replace(sanitizedURL, ">", "&gt;", -1)
		sanitizedURL = strings.Replace(sanitizedURL, "\"", "&quot;", -1)
		sanitizedURL = strings.Replace(sanitizedURL, "'", "&#39;", -1)
		
		// Use the sanitized URL in the text replacement
		params.TextReplacements["text-url"] = sanitizedURL
	}

	return params, nil
}

// ListSVGsHandler returns a list of available SVGs
func (h *SVGHandler) ListSVGsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("List SVGs request from: %s", r.RemoteAddr)
	
	svgs, err := h.processor.ListAvailableSVGs()
	if err != nil {
		log.Printf("Error listing SVGs: %v", err)
		http.Error(w, fmt.Sprintf("Error listing SVGs: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d SVG files", len(svgs))
	w.Header().Set("Content-Type", "text/plain")
	for _, svg := range svgs {
		fmt.Fprintf(w, "%s\n", svg)
	}
}
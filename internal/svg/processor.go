package svg

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Processor handles SVG processing operations
type Processor struct {
	BasePath string
}

// NewProcessor creates a new SVG processor with the given base path for SVG files
func NewProcessor(basePath string) *Processor {
	return &Processor{
		BasePath: basePath,
	}
}

// SVGParams represents parameters for SVG customization
type SVGParams struct {
	// Text replacements map with ID -> new text
	TextReplacements map[string]string
	// Color replacements map with ID -> new color
	ColorReplacements map[string]string
	// Width of the SVG
	Width string
	// Height of the SVG
	Height string
}

// ProcessSVG loads an SVG file and modifies it according to parameters
func (p *Processor) ProcessSVG(svgName string, params SVGParams) ([]byte, error) {
	// Construct the path to the SVG file
	svgPath := filepath.Join(p.BasePath, svgName)
	
	// Check if the file exists
	if _, err := os.Stat(svgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("SVG file %s not found", svgName)
	}
	
	// Read the SVG file
	svgData, err := os.ReadFile(svgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SVG file: %w", err)
	}
	
	// Parse the SVG to modify it
	modifiedSVG, err := p.modifySVG(svgData, params)
	if err != nil {
		return nil, fmt.Errorf("failed to modify SVG: %w", err)
	}
	
	return modifiedSVG, nil
}

// modifySVG parses and modifies SVG content according to parameters
func (p *Processor) modifySVG(svgData []byte, params SVGParams) ([]byte, error) {
	svgString := string(svgData)
	
	// Step 1: Fix any duplicate xmlns attributes
	// Ensure only one xmlns attribute exists
	if strings.Count(svgString, `xmlns="http://www.w3.org/2000/svg"`) > 1 {
		// Remove all occurrences
		svgString = strings.Replace(svgString, `xmlns="http://www.w3.org/2000/svg"`, "", -1)
		// Add back one occurrence at the right position
		svgString = strings.Replace(svgString, "<svg", `<svg xmlns="http://www.w3.org/2000/svg"`, 1)
	}
	
	// Step 2: Handle width and height modifications
	if params.Width != "" {
		// Replace width attribute
		widthPattern := `width="[^"]*"`
		svgString = regexp.MustCompile(widthPattern).ReplaceAllString(svgString, `width="`+params.Width+`"`)
	}
	
	if params.Height != "" {
		// Replace height attribute
		heightPattern := `height="[^"]*"`
		svgString = regexp.MustCompile(heightPattern).ReplaceAllString(svgString, `height="`+params.Height+`"`)
	}
	
	// Step 3: Handle text replacements
	for elementID, newText := range params.TextReplacements {
		// Find the pattern: id="elementID"...>text</
		pattern := `id="` + elementID + `"[^>]*>[^<]*<`
		replacement := func(match string) string {
			// Extract the part before the text
			beforeText := match[:strings.LastIndex(match, ">")+1]
			// Return before + new text + closing bracket
			return beforeText + newText + "<"
		}
		svgString = regexp.MustCompile(pattern).ReplaceAllStringFunc(svgString, replacement)
	}
	
	// Step 4: Handle color replacements
	for elementID, newColor := range params.ColorReplacements {
		// Find elements with the specified ID and replace their fill color
		pattern := `id="` + elementID + `"[^>]*fill="[^"]*"`
		replacement := func(match string) string {
			// Replace just the fill attribute value
			return regexp.MustCompile(`fill="[^"]*"`).ReplaceAllString(match, `fill="`+newColor+`"`)
		}
		svgString = regexp.MustCompile(pattern).ReplaceAllStringFunc(svgString, replacement)
	}
	
	return []byte(svgString), nil
}

// ListAvailableSVGs returns a list of available SVG files
func (p *Processor) ListAvailableSVGs() ([]string, error) {
	entries, err := os.ReadDir(p.BasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SVG directory: %w", err)
	}
	
	var svgFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".svg") {
			svgFiles = append(svgFiles, entry.Name())
		}
	}
	
	return svgFiles, nil
}
package svg

import (
	"fmt"
	"log"
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
		log.Printf("Replacing text for element ID: %s with: %s", elementID, newText)
		
		// Dump current SVG content for debugging
		log.Printf("Current SVG content (first 200 chars): %s", svgString[:min(200, len(svgString))])
		
		// First try with the specific structure of our SVG that uses tspan elements
		// This is a very specific pattern for the exact structure of our example SVG
		tspanSpecificPattern := `(<text id="` + elementID + `"[^>]*>[ \t\n\r]*<tspan[^>]*>)[^<]*(</tspan>)`
		if matches := regexp.MustCompile(tspanSpecificPattern).FindStringSubmatch(svgString); len(matches) > 0 {
			log.Printf("Found specific tspan pattern for %s. Match: %s", elementID, matches[0])
			newSvgString := regexp.MustCompile(tspanSpecificPattern).ReplaceAllString(svgString, "${1}"+newText+"${2}")
			if newSvgString != svgString {
				log.Printf("Text replacement succeeded with specific tspan pattern")
				svgString = newSvgString
				continue
			}
		}
		
		// Fall back to more general patterns
		log.Printf("Trying more general patterns for element ID: %s", elementID)
		
		// Try a pattern for direct text content
		directPattern := `(<text id="` + elementID + `"[^>]*>)([^<]*)(</text>)`
		if matches := regexp.MustCompile(directPattern).FindStringSubmatch(svgString); len(matches) > 0 {
			log.Printf("Found direct text pattern for %s. Match: %s", elementID, matches[0])
			newSvgString := regexp.MustCompile(directPattern).ReplaceAllString(svgString, "${1}"+newText+"${3}")
			if newSvgString != svgString {
				log.Printf("Text replacement succeeded with direct pattern")
				svgString = newSvgString
				continue
			}
		}
		
		// Try one more pattern for nested elements that's common in SVGs
		log.Printf("Trying complex pattern for element ID: %s", elementID)
		svgBeforeComplexPattern := svgString // save for comparison
		
		// This hacky approach is more likely to work with real SVGs
		// We find the text element, extract its content, and make a targeted replacement
		re := regexp.MustCompile(`<text[^>]*id="` + elementID + `"[^>]*>(.*?)</text>`)
		matches := re.FindStringSubmatch(svgString)
		if len(matches) > 0 {
			log.Printf("Found text element with id=%s: %s", elementID, matches[0])
			
			// See if there's a tspan inside
			tspanContent := regexp.MustCompile(`<tspan[^>]*>(.*?)</tspan>`).FindStringSubmatch(matches[1])
			if len(tspanContent) > 0 {
				log.Printf("Found tspan content: %s", tspanContent[1])
				// Replace just the text content inside the tspan
				newTextElement := strings.Replace(matches[0], tspanContent[1], newText, 1)
				svgString = strings.Replace(svgString, matches[0], newTextElement, 1)
			} else {
				// No tspan, replace direct text content
				newTextElement := strings.Replace(matches[0], matches[1], newText, 1)
				svgString = strings.Replace(svgString, matches[0], newTextElement, 1)
			}
			
			if svgString != svgBeforeComplexPattern {
				log.Printf("Text replacement succeeded with complex pattern")
				continue
			}
		}
		
		log.Printf("WARNING: No pattern matched for %s", elementID)
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
	
	// Utility function to get min of two integers
	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	
	log.Printf("Final SVG preview (first 100 chars): %s", svgString[:min(100, len(svgString))])
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
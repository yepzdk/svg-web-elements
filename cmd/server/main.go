package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/svg-web-elements/internal/handlers"
)

func main() {
	// Determine base path for static files
	baseDir := getBaseDir()
	svgDir := filepath.Join(baseDir, "static", "svg")

	// Create our SVG handler
	svgHandler := handlers.NewSVGHandler(svgDir)

	// Setup routes
	http.Handle("/ui/", http.StripPrefix("/ui/", svgHandler))
	http.HandleFunc("/list", svgHandler.ListSVGsHandler)
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		svgName := r.URL.Query().Get("svg")
		if svgName == "" {
			http.Error(w, "Missing svg parameter", http.StatusBadRequest)
			return
		}
		
		svgPath := filepath.Join(svgDir, svgName)
		svgData, err := os.ReadFile(svgPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading SVG: %v", err), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
		<html>
		<head>
			<title>SVG Debug - %s</title>
			<style>
				body { font-family: system-ui, sans-serif; padding: 2rem; line-height: 1.5; }
				pre { background: #f1f1f1; padding: 1rem; overflow: auto; }
				.highlight { background: yellow; }
				.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 2rem; }
				.element { margin-bottom: 1rem; border: 1px solid #ddd; padding: 1rem; }
				.controls { margin-bottom: 2rem; }
			</style>
		</head>
		<body>
			<h1>SVG Debug for %s</h1>
			
			<div class="controls">
				<a href="/">&larr; Back to home</a>
				<p>Use this page to understand the structure of the SVG and how to modify it with query parameters.</p>
				<p>
					<strong>Scaling Examples:</strong>
					<a href="/ui/%s?width=400" target="_blank">width=400</a> |
					<a href="/ui/%s?height=200" target="_blank">height=200</a> |
					<a href="/ui/%s?width=500&height=250" target="_blank">width=500&height=250</a>
				</p>
			</div>
			
			<div class="grid">
				<div>
					<h2>SVG Preview</h2>
					<div style="border: 1px solid #ddd; padding: 1rem; margin-bottom: 1rem;">
						%s
					</div>
					<p>To customize this SVG, use query parameters like:</p>
					<ul>
						<li><code>/ui/%s?text.text-title=Custom+Title</code></li>
						<li><code>/ui/%s?text.text-url=example.com</code></li>
						<li><code>/ui/%s?width=500&height=300</code></li>
						<li><code>/ui/%s?width=300</code> (height scales proportionally)</li>
						<li><code>/ui/%s?height=200</code> (width scales proportionally)</li>
					</ul>
				</div>
				
				<div>
					<h2>Text Elements</h2>
					<div id="text-elements">Loading...</div>
					
					<h2>Color Elements</h2>
					<div id="color-elements">Loading...</div>
					
					<h2>Scaling</h2>
					<div class="element">
						<strong>Original Size:</strong> <span id="original-size">Loading...</span><br>
						<strong>ViewBox:</strong> <span id="viewbox">Loading...</span><br>
						<p>The SVG will scale proportionally by default. You can specify either width or height (or both).</p>
					</div>
				</div>
			</div>
			
			<h2>Raw SVG Source</h2>
			<pre>%s</pre>
			
			<script>
			// Function to extract elements with IDs and fills/text content
			function analyzeSVG() {
				const parser = new DOMParser();
				const svgElement = document.querySelector('svg');
				const svgDoc = parser.parseFromString(svgElement.outerHTML, "image/svg+xml");
				
				// Get SVG dimensions and viewBox
				document.getElementById('original-size').textContent = 
					svgElement.getAttribute('width') + ' x ' + svgElement.getAttribute('height');
				document.getElementById('viewbox').textContent = 
					svgElement.getAttribute('viewBox') || 'Not specified';
				
				// Find all elements with IDs
				const allElements = svgDoc.querySelectorAll('[id]');
				let textHTML = '';
				let colorHTML = '';
				
				// Process all elements
				allElements.forEach(el => {
					// Check for text elements
					const textContent = getElementTextContent(el);
					if (textContent) {
						textHTML += '<div class="element">';
						textHTML += '<strong>ID:</strong> ' + el.id + '<br>';
						textHTML += '<strong>Text:</strong> "' + textContent + '"<br>';
						textHTML += '<strong>Element Type:</strong> ' + el.tagName + '<br>';
						textHTML += '<strong>Usage:</strong> <code>text.' + el.id + '=New+Text</code>';
						textHTML += '</div>';
					}
					
					// Check for elements with fill attributes
					if (el.getAttribute('fill')) {
						const fillColor = el.getAttribute('fill');
						colorHTML += '<div class="element">';
						colorHTML += '<strong>ID:</strong> ' + el.id + '<br>';
						colorHTML += '<strong>Element Type:</strong> ' + el.tagName + '<br>';
						colorHTML += '<strong>Current Color:</strong> <span style="display:inline-block;width:20px;height:20px;background:' + fillColor + '"></span> ' + fillColor + '<br>';
						colorHTML += '<strong>Usage:</strong> <code>color.' + el.id + '=%23ff0000</code> (for red)';
						colorHTML += '</div>';
					}
					
					// For elements that might accept fill but don't have it yet
					if (!el.getAttribute('fill') && (el.tagName === 'rect' || el.tagName === 'path' || 
						el.tagName === 'circle' || el.tagName === 'polygon' || el.tagName === 'g')) {
						colorHTML += '<div class="element">';
						colorHTML += '<strong>ID:</strong> ' + el.id + '<br>';
						colorHTML += '<strong>Element Type:</strong> ' + el.tagName + '<br>';
						colorHTML += '<strong>No Fill Attribute</strong> - Can be added with: <code>color.' + el.id + '=%23ff0000</code>';
						colorHTML += '</div>';
					}
				});
				
				// Helper function to get text content including from nested tspan elements
				function getElementTextContent(element) {
					if (element.tagName === 'text') {
						// For text elements, include text from child nodes
						return element.textContent.trim();
					} else if (element.querySelector('text')) {
						// For groups that contain text elements
						const textEl = element.querySelector('text');
						return textEl.textContent.trim();
					}
					// For other elements with text content
					return element.textContent.trim() || null;
				}
				
				document.getElementById('text-elements').innerHTML = textHTML || 'No text elements found';
				document.getElementById('color-elements').innerHTML = colorHTML || 'No color elements found';
			}
			
			// Run analysis when page loads
			window.onload = analyzeSVG;
			</script>
		</body>
		</html>`, svgName, svgName, svgName, svgName, svgName, string(svgData), svgName, svgName, svgName, svgName, svgName, 
		strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(svgData), "&", "&amp;"), "<", "&lt;"), ">", "&gt;"))
	})

	// Add a simple index page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>SVG Web Elements</title>
			<style>
				body {
					font-family: system-ui, -apple-system, sans-serif;
					max-width: 800px;
					margin: 0 auto;
					padding: 2rem;
					line-height: 1.5;
				}
				code {
					background: #f1f1f1;
					padding: 0.2rem 0.4rem;
					border-radius: 3px;
				}
				.example {
					margin: 2rem 0;
					padding: 1rem;
					border: 1px solid #e0e0e0;
					border-radius: 4px;
				}
				h3 {
					margin-top: 0;
				}
			</style>
		</head>
		<body>
			<h1>SVG Web Elements Service</h1>
			<p>This service provides customizable SVG illustrations via URL parameters.</p>
			
			<h2>Available SVGs</h2>
			<p>Check the <a href="/list">list of available SVGs</a>.</p>
			
			<h2>Usage</h2>
			<p>You can use the SVGs in your applications by creating an image tag with the URL:</p>
			<code>&lt;img src="https://this-service/ui/basic-auth.svg?width=400&amp;height=200&amp;text.text-title=Login" /&gt;</code>
			
			<h2>Diagnostics & Examples</h2>
			<p>Use the <a href="/debug?svg=basic-auth.svg">SVG diagnostic tool</a> to inspect SVG elements and their IDs.</p>
			
			<div class="example">
				<h3>Basic Example</h3>
				<img src="/ui/basic-auth.svg" alt="Basic Auth SVG" />
				<p>URL: <code>/ui/basic-auth.svg</code></p>
			</div>
			
			<div class="example">
				<h3>Modified Text</h3>
				<img src="/ui/basic-auth.svg?text.text-title=Login&text.text-url=example.com" alt="Modified Text SVG" />
				<p>URL: <code>/ui/basic-auth.svg?text.text-title=Login&amp;text.text-url=example.com</code></p>
			</div>
			
			<div class="example">
				<h3>Modified Size (Proportional Scaling)</h3>
				<img src="/ui/basic-auth.svg?width=400&height=200" alt="Modified Size SVG" />
				<p>URL: <code>/ui/basic-auth.svg?width=400&amp;height=200</code></p>
				<p><small>You can specify just width or height, and the other dimension will scale proportionally.</small></p>
			</div>
			
			<div class="example">
				<h3>Modified Colors</h3>
				<img src="/ui/basic-auth.svg?color.page-background=%23f0f9ff&color.prompt-background=%23ffffff&color.btn-background_2=%230ea5e9" alt="Modified Colors SVG" />
				<p>URL: <code>/ui/basic-auth.svg?color.page-background=%23f0f9ff&amp;color.prompt-background=%23ffffff&amp;color.btn-background_2=%230ea5e9</code></p>
				<p><small>Note: Use <code>%23</code> instead of <code>#</code> in URLs for hex colors</small></p>
			</div>
			
			<h2>Parameters</h2>
			<ul>
				<li><code>width</code> - Set the SVG width (other dimension scales proportionally if height not specified)</li>
				<li><code>height</code> - Set the SVG height (other dimension scales proportionally if width not specified)</li>
				<li><code>text.{element-id}</code> - Replace text in element with ID</li>
				<li><code>color.{element-id}</code> - Change color of element with ID (use <code>%23</code> instead of <code>#</code> for hex colors)</li>
				<li><code>url</code> - External URL to use (shown in the URL field)</li>
			</ul>
			<p>Try the <a href="/debug?svg=basic-auth.svg">SVG debug tool</a> to see all available element IDs.</p>
			
			<footer>
				<p>SVG Web Elements Service</p>
			</footer>
		</body>
		</html>
		`)
	})

	// Start the server
	port := getEnv("PORT", "8082")
	log.Printf("Starting server on port %s...", port)
	log.Printf("SVG files will be served from: %s", svgDir)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// getBaseDir returns the base directory for the application
func getBaseDir() string {
	// Try to use working directory first
	wd, err := os.Getwd()
	if err == nil {
		// Check if we're running from the cmd/server directory
		if filepath.Base(wd) == "server" && filepath.Base(filepath.Dir(wd)) == "cmd" {
			return filepath.Dir(filepath.Dir(wd))
		}
		return wd
	}
	
	// Fallback to executable directory
	ex, err := os.Executable()
	if err == nil {
		return filepath.Dir(ex)
	}
	
	// Last resort
	return "."
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
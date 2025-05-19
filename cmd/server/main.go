package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
			
			<h2>Examples</h2>
			
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
				<h3>Modified Size</h3>
				<img src="/ui/basic-auth.svg?width=400&height=200" alt="Modified Size SVG" />
				<p>URL: <code>/ui/basic-auth.svg?width=400&amp;height=200</code></p>
			</div>
			
			<div class="example">
				<h3>Modified Colors</h3>
				<img src="/ui/basic-auth.svg?color.page-background=#f0f9ff&color.btn-background_2=#0ea5e9" alt="Modified Colors SVG" />
				<p>URL: <code>/ui/basic-auth.svg?color.page-background=#f0f9ff&amp;color.btn-background_2=#0ea5e9</code></p>
			</div>
			
			<h2>Parameters</h2>
			<ul>
				<li><code>width</code> - Set the SVG width</li>
				<li><code>height</code> - Set the SVG height</li>
				<li><code>text.{element-id}</code> - Replace text in element with ID</li>
				<li><code>color.{element-id}</code> - Change color of element with ID</li>
				<li><code>url</code> - External URL to use (shown in the URL field)</li>
			</ul>
			
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
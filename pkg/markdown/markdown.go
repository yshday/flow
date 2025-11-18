package markdown

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

// Renderer handles markdown rendering and HTML sanitization
type Renderer struct {
	sanitizer *bluemonday.Policy
}

// NewRenderer creates a new markdown renderer
func NewRenderer() *Renderer {
	// Create HTML sanitizer policy (allows safe HTML tags)
	sanitizer := bluemonday.UGCPolicy()

	return &Renderer{
		sanitizer: sanitizer,
	}
}

// RenderToHTML converts markdown to sanitized HTML
func (r *Renderer) RenderToHTML(md string) string {
	// Create markdown parser with common extensions (must create new parser for each call)
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// Parse markdown
	doc := p.Parse([]byte(md))

	// Create HTML renderer
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// Render to HTML
	unsafeHTML := markdown.Render(doc, renderer)

	// Sanitize HTML to prevent XSS
	safeHTML := r.sanitizer.SanitizeBytes(unsafeHTML)

	return string(safeHTML)
}

// StripHTML removes all HTML tags and returns plain text
func (r *Renderer) StripHTML(html string) string {
	return bluemonday.StrictPolicy().Sanitize(html)
}

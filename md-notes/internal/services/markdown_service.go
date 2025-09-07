package services
import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
)

// ConvertMarkdownToHTML converts markdown text to HTML string
func ConvertMarkdownToHTML(md string) string {
	renderer := html.NewRenderer(html.RendererOptions{})
	return string(markdown.ToHTML([]byte(md), nil, renderer))
}
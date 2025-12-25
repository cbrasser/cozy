package ebook

import (
	"strings"

	"github.com/cbrasser/cozy/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"golang.org/x/net/html"
)

// Renderer converts HTML to styled terminal text
type Renderer struct {
	theme *config.Theme
	width int
}

// NewRenderer creates a new HTML renderer
func NewRenderer(theme *config.Theme, width int) *Renderer {
	return &Renderer{
		theme: theme,
		width: width,
	}
}

// Render converts HTML to styled text
func (r *Renderer) Render(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// Fallback to simple text stripping
		return htmlToText(htmlContent)
	}

	var result strings.Builder
	r.renderNode(doc, &result, &renderContext{})

	return strings.TrimSpace(result.String())
}

// renderContext tracks the current rendering state
type renderContext struct {
	inHeading    int  // 0 = none, 1-6 = h1-h6
	inBlockquote bool
	inPre        bool
	inCode       bool
	inEmphasis   bool
	inStrong     bool
	listLevel    int
}

// clone creates a copy of the context
func (ctx *renderContext) clone() *renderContext {
	newCtx := *ctx
	return &newCtx
}

// renderNode recursively renders an HTML node
func (r *Renderer) renderNode(n *html.Node, out *strings.Builder, ctx *renderContext) {
	switch n.Type {
	case html.TextNode:
		text := n.Data

		// Preserve whitespace in <pre> tags
		if !ctx.inPre {
			text = strings.TrimSpace(text)
		}

		if text != "" {
			r.writeStyledText(out, text, ctx)
		}

	case html.ElementNode:
		r.renderElement(n, out, ctx)

	case html.DocumentNode:
		// Process all children of document
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			r.renderNode(c, out, ctx)
		}
	}
}

// renderElement renders an HTML element
func (r *Renderer) renderElement(n *html.Node, out *strings.Builder, ctx *renderContext) {
	newCtx := ctx.clone()

	// Handle element-specific behavior
	switch n.Data {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		out.WriteString("\n\n")
		switch n.Data {
		case "h1":
			newCtx.inHeading = 1
		case "h2":
			newCtx.inHeading = 2
		case "h3":
			newCtx.inHeading = 3
		case "h4":
			newCtx.inHeading = 4
		case "h5":
			newCtx.inHeading = 5
		case "h6":
			newCtx.inHeading = 6
		}

	case "p":
		out.WriteString("\n\n")

	case "blockquote":
		out.WriteString("\n\n")
		newCtx.inBlockquote = true

	case "pre":
		out.WriteString("\n\n")
		newCtx.inPre = true
		newCtx.inCode = true

	case "code":
		if !ctx.inPre {
			newCtx.inCode = true
		}

	case "em", "i":
		newCtx.inEmphasis = true

	case "strong", "b":
		newCtx.inStrong = true

	case "br":
		out.WriteString("\n")
		return

	case "hr":
		out.WriteString("\n\n")
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.MutedTextColor))
		out.WriteString(style.Render(strings.Repeat("─", min(r.width, 80))))
		out.WriteString("\n\n")
		return

	case "ul", "ol":
		out.WriteString("\n")
		newCtx.listLevel++

	case "li":
		indent := strings.Repeat("  ", ctx.listLevel-1)
		out.WriteString("\n" + indent + "• ")

	case "div", "span", "a":
		// Pass through, just render children
	}

	// Render children with new context
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		r.renderNode(c, out, newCtx)
	}

	// Post-element formatting
	switch n.Data {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		out.WriteString("\n")
	case "blockquote", "pre":
		out.WriteString("\n")
	case "ul", "ol":
		out.WriteString("\n")
	}
}

// writeStyledText applies styling and writes text
func (r *Renderer) writeStyledText(out *strings.Builder, text string, ctx *renderContext) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.TextColor))

	// Calculate effective width (accounting for borders and padding)
	effectiveWidth := r.width
	if effectiveWidth <= 0 {
		effectiveWidth = 80
	}

	// Apply context-specific styling
	if ctx.inHeading > 0 {
		style = style.
			Foreground(lipgloss.Color(r.theme.HeadingColor)).
			Bold(true)

		if ctx.inHeading == 1 {
			style = style.Underline(true)
		}

		// Add heading prefix
		prefix := strings.Repeat("#", ctx.inHeading) + " "
		text = prefix + text

		// Wrap heading text
		text = wordwrap.String(text, effectiveWidth)
	}

	if ctx.inBlockquote {
		// Wrap text before styling (account for border + padding = 4 chars)
		wrappedText := wordwrap.String(text, max(effectiveWidth-4, 40))

		// Format blockquote with left border
		lines := strings.Split(wrappedText, "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			quoteStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(r.theme.QuoteColor)).
				Italic(true).
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color(r.theme.QuoteBorderColor)).
				PaddingLeft(1)

			out.WriteString(quoteStyle.Render(line))
			if i < len(lines)-1 {
				out.WriteString("\n")
			}
		}
		return
	}

	if ctx.inCode {
		style = style.
			Foreground(lipgloss.Color(r.theme.CodeTextColor)).
			Background(lipgloss.Color(r.theme.CodeBgColor))

		if ctx.inPre {
			style = style.Padding(0, 1)
			// Wrap code blocks (account for padding = 2 chars)
			text = wordwrap.String(text, max(effectiveWidth-2, 40))
		} else {
			style = style.Padding(0, 1)
		}
	} else {
		// Wrap regular text
		text = wordwrap.String(text, effectiveWidth)

		// Apply inline formatting
		if ctx.inEmphasis {
			style = style.
				Foreground(lipgloss.Color(r.theme.EmphasisColor)).
				Italic(true)
		}

		if ctx.inStrong {
			style = style.
				Foreground(lipgloss.Color(r.theme.StrongColor)).
				Bold(true)
		}
	}

	out.WriteString(style.Render(text))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RenderToStyledText is the main entry point for rendering HTML
func RenderToStyledText(htmlContent string, theme *config.Theme, width int) string {
	renderer := NewRenderer(theme, width)
	result := renderer.Render(htmlContent)

	// If rendering produced no output, fall back to simple text extraction
	if strings.TrimSpace(result) == "" {
		return htmlToText(htmlContent)
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Debug helper - simple text extraction for testing
func ExtractPlainText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlToText(htmlContent)
	}

	var result strings.Builder
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				result.WriteString(text)
				result.WriteString(" ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)
	return strings.TrimSpace(result.String())
}

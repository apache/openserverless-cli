package tools

import (
	"regexp"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/jaytaylor/html2text"
)

func ExtractUsage(input string) string {
	// Regular expression to match the block of text
	re := regexp.MustCompile("(?s)```text\n(.*?Usage:.*?[^`]*)```")

	// Find the match
	matches := re.FindStringSubmatch(input)

	// Return the match, or an empty string if no match is found
	if len(matches) > 0 {
		return matches[1]
	}
	return "Cannot find an Usage: block"
}

func MarkdownToText(input string) string {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(input))
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	markdown := markdown.Render(doc, renderer)

	text, err := html2text.FromString(string(markdown), html2text.Options{PrettyTables: true})
	if err != nil {
		return err.Error()
	}
	return text
}

package tools

import (
	"fmt"
)

func Example_markdown2text() {
	markdown, _ := GetMarkDown("base64")
	fmt.Println(len(MarkdownToText(markdown)))
	// Output: 737
}

func Example_extractUsage() {
	markdown, _ := GetMarkDown("base64")
	fmt.Println(len(ExtractUsage(markdown)))
	// Output: 40
}

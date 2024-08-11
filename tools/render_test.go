package tools

import "fmt"

func Example_markdown2text() {
	fmt.Println(len(MarkdownToText(base64usage)))
	// Output: 945
}

func Example_extractUsage() {
	fmt.Println(len(ExtractUsage(base64usage)))
	// Output: 40

}

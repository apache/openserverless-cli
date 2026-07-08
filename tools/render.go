// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

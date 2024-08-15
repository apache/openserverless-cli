package tools

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMarkdownHelp(t *testing.T) {
	for _, s := range ToolList {
		if s.HasDoc {
			opt := MarkdownHelp(s.Name)
			if opt == "" {
				t.Fatalf("Tool %s doesn't have valid help", s.Name)
			}
		}
	}
}

func TestGetMarkDownSuccess(t *testing.T) {
	t.Helper()
	_, err := GetMarkDown("base64")
	require.NoError(t, err)
}

func TestGetMarkDownError(t *testing.T) {
	t.Helper()
	_, err := GetMarkDown("notexistenttool")
	require.Error(t, err)
}

func ExampleMergeToolsList() {
	mergedList := MergeToolsList([]string{})
	fmt.Println(len(mergedList))
	//Output: 24
}

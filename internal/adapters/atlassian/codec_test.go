package atlassian

import (
	"encoding/json"
	"testing"

	adatamodels "github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/go-faster/errors"
)

func TestMarkdownToADF_EmptyString(t *testing.T) {
	result := MarkdownToADF("")
	if result.Type != "doc" || result.Version != 1 {
		t.Error(errors.Errorf("expected doc node, got %+v", result))
	}
	if len(result.Content) != 0 {
		t.Error(errors.Errorf("expected no content for empty string, got %d nodes", len(result.Content)))
	}
}

func TestMarkdownToADF_PlainText(t *testing.T) {
	result := MarkdownToADF("Hello world")
	assertDocWithParagraph(t, result)
	paragraph := result.Content[0]
	if len(paragraph.Content) != 1 || paragraph.Content[0].Text != "Hello world" {
		t.Error(errors.Errorf("expected text 'Hello world', got %+v", paragraph.Content))
	}
}

func TestMarkdownToADF_Headings(t *testing.T) {
	tests := []struct {
		input string
		level float64
	}{
		{"# H1", 1},
		{"## H2", 2},
		{"### H3", 3},
		{"#### H4", 4},
		{"##### H5", 5},
		{"###### H6", 6},
	}
	for _, tt := range tests {
		result := MarkdownToADF(tt.input)
		if len(result.Content) == 0 {
			t.Fatal(errors.Errorf("no content for %q", tt.input))
		}
		heading := result.Content[0]
		if heading.Type != "heading" {
			t.Error(errors.Errorf("expected heading, got %s for %q", heading.Type, tt.input))
		}
		if heading.Attrs["level"] != tt.level {
			t.Error(errors.Errorf("expected level %v, got %v for %q", tt.level, heading.Attrs["level"], tt.input))
		}
	}
}

func TestMarkdownToADF_Bold(t *testing.T) {
	result := MarkdownToADF("**bold**")
	assertMark(t, result.Content[0].Content[0], "bold", "strong")
}

func TestMarkdownToADF_Italic(t *testing.T) {
	result := MarkdownToADF("*italic*")
	assertMark(t, result.Content[0].Content[0], "italic", "em")
}

func TestMarkdownToADF_InlineCode(t *testing.T) {
	result := MarkdownToADF("`code`")
	assertMark(t, result.Content[0].Content[0], "code", "code")
}

func TestMarkdownToADF_Strikethrough(t *testing.T) {
	result := MarkdownToADF("~~strike~~")
	assertMark(t, result.Content[0].Content[0], "strike", "strike")
}

func TestMarkdownToADF_Link(t *testing.T) {
	result := MarkdownToADF("[click](https://example.com)")
	paragraph := result.Content[0]
	if len(paragraph.Content) == 0 {
		t.Fatal(errors.New("no content in paragraph"))
	}
	textNode := paragraph.Content[0]
	if textNode.Text != "click" {
		t.Error(errors.Errorf("expected text 'click', got %q", textNode.Text))
	}
	found := false
	for _, mark := range textNode.Marks {
		if mark.Type == "link" {
			found = true
			if mark.Attrs["href"] != "https://example.com" {
				t.Error(errors.Errorf("expected href https://example.com, got %v", mark.Attrs["href"]))
			}
		}
	}
	if !found {
		t.Error(errors.New("expected link mark not found"))
	}
}

func TestMarkdownToADF_BulletList(t *testing.T) {
	result := MarkdownToADF("- item 1\n- item 2\n- item 3")
	list := result.Content[0]
	if list.Type != "bulletList" {
		t.Error(errors.Errorf("expected bulletList, got %s", list.Type))
	}
	if len(list.Content) != 3 {
		t.Error(errors.Errorf("expected 3 list items, got %d", len(list.Content)))
	}
}

func TestMarkdownToADF_OrderedList(t *testing.T) {
	result := MarkdownToADF("1. first\n2. second")
	list := result.Content[0]
	if list.Type != "orderedList" {
		t.Error(errors.Errorf("expected orderedList, got %s", list.Type))
	}
	if len(list.Content) != 2 {
		t.Error(errors.Errorf("expected 2 list items, got %d", len(list.Content)))
	}
}

func TestMarkdownToADF_FencedCodeBlock(t *testing.T) {
	result := MarkdownToADF("```go\nfunc main() {}\n```")
	codeBlock := result.Content[0]
	if codeBlock.Type != "codeBlock" {
		t.Error(errors.Errorf("expected codeBlock, got %s", codeBlock.Type))
	}
	if codeBlock.Attrs["language"] != "go" {
		t.Error(errors.Errorf("expected language 'go', got %v", codeBlock.Attrs["language"]))
	}
	if len(codeBlock.Content) == 0 || codeBlock.Content[0].Text != "func main() {}\n" {
		t.Error(errors.Errorf("unexpected code content: %+v", codeBlock.Content))
	}
}

func TestMarkdownToADF_Blockquote(t *testing.T) {
	result := MarkdownToADF("> quoted text")
	blockquote := result.Content[0]
	if blockquote.Type != "blockquote" {
		t.Error(errors.Errorf("expected blockquote, got %s", blockquote.Type))
	}
	if len(blockquote.Content) == 0 || blockquote.Content[0].Type != "paragraph" {
		t.Error(errors.Errorf("expected paragraph inside blockquote, got %+v", blockquote.Content))
	}
}

func TestMarkdownToADF_HorizontalRule(t *testing.T) {
	result := MarkdownToADF("above\n\n---\n\nbelow")
	found := false
	for _, node := range result.Content {
		if node.Type == "rule" {
			found = true
			break
		}
	}
	if !found {
		t.Error(errors.New("expected rule node not found"))
	}
}

func TestMarkdownToADF_MixedContent(t *testing.T) {
	result := MarkdownToADF("# Title\n\nSome **bold** text.\n\n- item 1\n- item 2\n\n```js\nconsole.log('hi')\n```")
	if len(result.Content) < 4 {
		t.Fatal(errors.Errorf("expected at least 4 nodes, got %d", len(result.Content)))
	}
	if result.Content[0].Type != "heading" || result.Content[1].Type != "paragraph" || result.Content[2].Type != "bulletList" || result.Content[3].Type != "codeBlock" {
		t.Error(errors.Errorf("unexpected mixed content layout: %+v", result.Content))
	}
}

func TestMarkdownToADF_BoldItalic(t *testing.T) {
	result := MarkdownToADF("***bold italic***")
	textNode := result.Content[0].Content[0]
	hasStrong := false
	hasEm := false
	for _, mark := range textNode.Marks {
		if mark.Type == "strong" {
			hasStrong = true
		}
		if mark.Type == "em" {
			hasEm = true
		}
	}
	if !hasStrong || !hasEm {
		t.Error(errors.Errorf("expected both strong and em marks, got %+v", textNode.Marks))
	}
}

func TestMarkdownToADF_ValidJSON(t *testing.T) {
	result := MarkdownToADF("# Hello\n\nWorld **bold** and *italic*\n\n- list\n\n```go\ncode\n```")
	if _, err := json.Marshal(result); err != nil {
		t.Fatal(errors.Wrap(err, "failed to marshal to JSON"))
	}
}

func assertDocWithParagraph(t *testing.T, doc *adatamodels.CommentNodeScheme) {
	t.Helper()
	if doc.Type != "doc" {
		t.Fatal(errors.Errorf("expected doc, got %s", doc.Type))
	}
	if len(doc.Content) == 0 || doc.Content[0].Type != "paragraph" {
		t.Fatal(errors.Errorf("expected paragraph, got %+v", doc.Content))
	}
}

func assertMark(t *testing.T, node *adatamodels.CommentNodeScheme, expectedText, markType string) {
	t.Helper()
	if node.Text != expectedText {
		t.Error(errors.Errorf("expected text %q, got %q", expectedText, node.Text))
	}
	found := false
	for _, mark := range node.Marks {
		if mark.Type == markType {
			found = true
		}
	}
	if !found {
		t.Error(errors.Errorf("expected mark %q not found in %+v", markType, node.Marks))
	}
}

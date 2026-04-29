package atlassian

import (
	"bytes"
	"fmt"
	"strings"

	adatamodels "github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func MarkdownToADF(input string) *adatamodels.CommentNodeScheme {
	doc := &adatamodels.CommentNodeScheme{
		Version: 1,
		Type:    "doc",
	}

	if input == "" {
		return doc
	}

	source := []byte(input)
	md := goldmark.New(
		goldmark.WithExtensions(extension.Strikethrough),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	reader := text.NewReader(source)
	tree := md.Parser().Parse(reader)
	walkChildren(tree, doc, source)
	return doc
}

func RenderADF(node *adatamodels.CommentNodeScheme) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	renderADFNode(node, &builder, 0, "")
	return strings.TrimSpace(builder.String())
}

func walkChildren(parent ast.Node, adfParent *adatamodels.CommentNodeScheme, source []byte) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		nodes := convertNode(child, source)
		for _, node := range nodes {
			adfParent.AppendNode(node)
		}
	}
}

func convertNode(n ast.Node, source []byte) []*adatamodels.CommentNodeScheme {
	switch node := n.(type) {
	case *ast.Paragraph:
		paragraph := &adatamodels.CommentNodeScheme{Type: "paragraph"}
		walkInline(node, paragraph, source, nil)
		return []*adatamodels.CommentNodeScheme{paragraph}
	case *ast.Heading:
		heading := &adatamodels.CommentNodeScheme{
			Type:  "heading",
			Attrs: map[string]any{"level": float64(node.Level)},
		}
		walkInline(node, heading, source, nil)
		return []*adatamodels.CommentNodeScheme{heading}
	case *ast.FencedCodeBlock:
		language := ""
		if node.Language(source) != nil {
			language = string(node.Language(source))
		}
		codeBlock := &adatamodels.CommentNodeScheme{Type: "codeBlock"}
		if language != "" {
			codeBlock.Attrs = map[string]any{"language": language}
		}
		var buffer bytes.Buffer
		for index := 0; index < node.Lines().Len(); index++ {
			line := node.Lines().At(index)
			buffer.Write(line.Value(source))
		}
		if buffer.Len() > 0 {
			codeBlock.AppendNode(&adatamodels.CommentNodeScheme{Type: "text", Text: buffer.String()})
		}
		return []*adatamodels.CommentNodeScheme{codeBlock}
	case *ast.CodeBlock:
		codeBlock := &adatamodels.CommentNodeScheme{Type: "codeBlock"}
		var buffer bytes.Buffer
		for index := 0; index < node.Lines().Len(); index++ {
			line := node.Lines().At(index)
			buffer.Write(line.Value(source))
		}
		if buffer.Len() > 0 {
			codeBlock.AppendNode(&adatamodels.CommentNodeScheme{Type: "text", Text: buffer.String()})
		}
		return []*adatamodels.CommentNodeScheme{codeBlock}
	case *ast.Blockquote:
		blockquote := &adatamodels.CommentNodeScheme{Type: "blockquote"}
		walkChildren(node, blockquote, source)
		return []*adatamodels.CommentNodeScheme{blockquote}
	case *ast.List:
		listType := "bulletList"
		if node.IsOrdered() {
			listType = "orderedList"
		}
		list := &adatamodels.CommentNodeScheme{Type: listType}
		if node.IsOrdered() && node.Start != 1 {
			list.Attrs = map[string]any{"order": float64(node.Start)}
		}
		walkChildren(node, list, source)
		return []*adatamodels.CommentNodeScheme{list}
	case *ast.ListItem:
		item := &adatamodels.CommentNodeScheme{Type: "listItem"}
		walkChildren(node, item, source)
		return []*adatamodels.CommentNodeScheme{item}
	case *ast.ThematicBreak:
		return []*adatamodels.CommentNodeScheme{{Type: "rule"}}
	case *ast.TextBlock:
		paragraph := &adatamodels.CommentNodeScheme{Type: "paragraph"}
		walkInline(node, paragraph, source, nil)
		return []*adatamodels.CommentNodeScheme{paragraph}
	default:
		var result []*adatamodels.CommentNodeScheme
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			result = append(result, convertNode(child, source)...)
		}
		return result
	}
}

func walkInline(parent ast.Node, adfParent *adatamodels.CommentNodeScheme, source []byte, marks []*adatamodels.MarkScheme) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		convertInline(child, adfParent, source, marks)
	}
}

func convertInline(n ast.Node, adfParent *adatamodels.CommentNodeScheme, source []byte, marks []*adatamodels.MarkScheme) {
	switch node := n.(type) {
	case *ast.Text:
		text := string(node.Segment.Value(source))
		if text != "" {
			textNode := &adatamodels.CommentNodeScheme{Type: "text", Text: text}
			if len(marks) > 0 {
				textNode.Marks = copyMarks(marks)
			}
			adfParent.AppendNode(textNode)
		}
		if node.HardLineBreak() {
			adfParent.AppendNode(&adatamodels.CommentNodeScheme{Type: "hardBreak"})
		}
	case *ast.String:
		text := string(node.Value)
		if text != "" {
			textNode := &adatamodels.CommentNodeScheme{Type: "text", Text: text}
			if len(marks) > 0 {
				textNode.Marks = copyMarks(marks)
			}
			adfParent.AppendNode(textNode)
		}
	case *ast.Emphasis:
		markType := "em"
		if node.Level == 2 {
			markType = "strong"
		}
		walkInline(node, adfParent, source, append(copyMarks(marks), &adatamodels.MarkScheme{Type: markType}))
	case *ast.CodeSpan:
		var buffer bytes.Buffer
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if textNode, ok := child.(*ast.Text); ok {
				buffer.Write(textNode.Segment.Value(source))
			}
		}
		if buffer.Len() > 0 {
			adfParent.AppendNode(&adatamodels.CommentNodeScheme{
				Type:  "text",
				Text:  buffer.String(),
				Marks: append(copyMarks(marks), &adatamodels.MarkScheme{Type: "code"}),
			})
		}
	case *ast.Link:
		walkInline(node, adfParent, source, append(copyMarks(marks), &adatamodels.MarkScheme{
			Type:  "link",
			Attrs: map[string]any{"href": string(node.Destination)},
		}))
	case *ast.AutoLink:
		url := string(node.URL(source))
		adfParent.AppendNode(&adatamodels.CommentNodeScheme{
			Type: "text",
			Text: url,
			Marks: append(copyMarks(marks), &adatamodels.MarkScheme{
				Type:  "link",
				Attrs: map[string]any{"href": url},
			}),
		})
	case *east.Strikethrough:
		walkInline(node, adfParent, source, append(copyMarks(marks), &adatamodels.MarkScheme{Type: "strike"}))
	default:
		walkInline(n, adfParent, source, marks)
	}
}

func copyMarks(marks []*adatamodels.MarkScheme) []*adatamodels.MarkScheme {
	if len(marks) == 0 {
		return nil
	}
	result := make([]*adatamodels.MarkScheme, len(marks))
	copy(result, marks)
	return result
}

func renderADFNode(node *adatamodels.CommentNodeScheme, builder *strings.Builder, depth int, listPrefix string) {
	if node == nil {
		return
	}

	switch node.Type {
	case "doc":
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
		}
	case "paragraph":
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
		}
		builder.WriteString("\n\n")
	case "text":
		text := node.Text
		for _, mark := range node.Marks {
			switch mark.Type {
			case "strong":
				text = "**" + text + "**"
			case "em":
				text = "*" + text + "*"
			case "code":
				text = "`" + text + "`"
			case "strike":
				text = "~~" + text + "~~"
			case "underline":
				text = "__" + text + "__"
			}
		}
		builder.WriteString(text)
	case "hardBreak":
		builder.WriteByte('\n')
	case "heading":
		level := 1
		if node.Attrs != nil {
			if value, ok := node.Attrs["level"].(float64); ok {
				level = int(value)
			}
		}
		builder.WriteString(strings.Repeat("#", level) + " ")
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
		}
		builder.WriteString("\n\n")
	case "bulletList":
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, "- ")
		}
	case "orderedList":
		for index, child := range node.Content {
			renderADFNode(child, builder, depth, fmt.Sprintf("%d. ", index+1))
		}
	case "listItem":
		if listPrefix != "" {
			builder.WriteString(strings.Repeat("  ", depth))
			builder.WriteString(listPrefix)
		}
		for _, child := range node.Content {
			renderADFNode(child, builder, depth+1, "")
		}
	case "codeBlock":
		language := ""
		if node.Attrs != nil {
			if value, ok := node.Attrs["language"].(string); ok {
				language = value
			}
		}
		builder.WriteString("```" + language + "\n")
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
		}
		builder.WriteString("```\n\n")
	case "blockquote":
		var inner strings.Builder
		for _, child := range node.Content {
			renderADFNode(child, &inner, depth, listPrefix)
		}
		for _, line := range strings.Split(strings.TrimSpace(inner.String()), "\n") {
			builder.WriteString("> " + line + "\n")
		}
		builder.WriteByte('\n')
	case "rule":
		builder.WriteString("---\n\n")
	case "table":
		builder.WriteString("\n[Table Content]\n")
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
		}
		builder.WriteByte('\n')
	case "tableRow":
		builder.WriteString("| ")
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
			builder.WriteString(" | ")
		}
		builder.WriteByte('\n')
	case "tableHeader", "tableCell", "mediaSingle", "mediaGroup":
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
		}
	case "media":
		if node.Attrs == nil {
			builder.WriteString("[Media/Image]")
			return
		}
		alt, _ := node.Attrs["alt"].(string)
		mediaType, _ := node.Attrs["type"].(string)
		mediaID, _ := node.Attrs["id"].(string)
		switch {
		case alt != "":
			builder.WriteString("[Media: " + alt)
		case mediaType != "":
			builder.WriteString("[Media: " + mediaType)
		default:
			builder.WriteString("[Media")
		}
		if width, ok := node.Attrs["width"].(float64); ok {
			if height, ok := node.Attrs["height"].(float64); ok {
				builder.WriteString(fmt.Sprintf(" (%dx%d)", int(width), int(height)))
			}
		}
		if mediaID != "" {
			builder.WriteString(" | id=" + mediaID)
		}
		builder.WriteString("]")
	case "mention":
		if node.Attrs != nil {
			if text, ok := node.Attrs["text"].(string); ok {
				builder.WriteString("@" + text)
			}
		}
	case "emoji":
		if node.Attrs != nil {
			if shortName, ok := node.Attrs["shortName"].(string); ok {
				builder.WriteString(shortName)
			}
		}
	case "inlineCard":
		if node.Attrs != nil {
			if url, ok := node.Attrs["url"].(string); ok {
				builder.WriteString(url)
			}
		}
	default:
		for _, child := range node.Content {
			renderADFNode(child, builder, depth, listPrefix)
		}
	}
}

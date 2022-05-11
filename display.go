package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/fatih/color"
	colorable "github.com/mattn/go-colorable"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func init() {
	color.Output = colorable.NewColorableStdout()
}

type Displayer interface {
	Display(io.Writer, []*html.Node)
}

func ParseDisplayer(cmd string) error {
	attrRe := regexp.MustCompile(`attr\{([a-zA-Z\-]+)\}`)
	if cmd == "text{}" {
		pupDisplayer = TextDisplayer{}
	} else if cmd == "json{}" {
		pupDisplayer = JSONDisplayer{}
	} else if match := attrRe.FindAllStringSubmatch(cmd, -1); len(match) == 1 {
		pupDisplayer = AttrDisplayer{
			Attr: match[0][1],
		}
	} else {
		return fmt.Errorf("Unknown displayer")
	}
	return nil
}

// Is this node a tag with no end tag such as <meta> or <br>?
// http://www.w3.org/TR/html-markup/syntax.html#syntax-elements
func isVoidElement(n *html.Node) bool {
	switch n.DataAtom {
	case atom.Area, atom.Base, atom.Br, atom.Col, atom.Command, atom.Embed,
		atom.Hr, atom.Img, atom.Input, atom.Keygen, atom.Link,
		atom.Meta, atom.Param, atom.Source, atom.Track, atom.Wbr:
		return true
	}
	return false
}

var (
	// Colors
	tagColor     *color.Color = color.New(color.FgCyan)
	tokenColor                = color.New(color.FgCyan)
	attrKeyColor              = color.New(color.FgMagenta)
	quoteColor                = color.New(color.FgBlue)
	commentColor              = color.New(color.FgYellow)
)

type TreeDisplayer struct {
}

func (t TreeDisplayer) Display(w io.Writer, nodes []*html.Node) {
	for _, node := range nodes {
		t.printNode(w, node, 0)
	}
}

// The <pre> tag indicates that the text within it should always be formatted
// as is. See https://github.com/ericchiang/pup/issues/33
func (t TreeDisplayer) printPre(w io.Writer, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		s := n.Data
		if pupEscapeHTML {
			// don't escape javascript
			if n.Parent == nil || n.Parent.DataAtom != atom.Script {
				s = html.EscapeString(s)
			}
		}
		fmt.Fprint(w, s)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.printPre(w, c)
		}
	case html.ElementNode:
		fmt.Fprintf(w, "<%s", n.Data)
		for _, a := range n.Attr {
			val := a.Val
			if pupEscapeHTML {
				val = html.EscapeString(val)
			}
			fmt.Fprintf(w, ` %s="%s"`, a.Key, val)
		}
		fmt.Fprint(w, ">")
		if !isVoidElement(n) {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				t.printPre(w, c)
			}
			fmt.Fprintf(w, "</%s>", n.Data)
		}
	case html.CommentNode:
		data := n.Data
		if pupEscapeHTML {
			data = html.EscapeString(data)
		}
		fmt.Fprintf(w, "<!--%s-->\n", data)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.printPre(w, c)
		}
	case html.DoctypeNode, html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			t.printPre(w, c)
		}
	}
}

// Print a node and all of it's children to `maxlevel`.
func (t TreeDisplayer) printNode(w io.Writer, n *html.Node, level int) {
	switch n.Type {
	case html.TextNode:
		s := n.Data
		if pupEscapeHTML {
			// don't escape javascript
			if n.Parent == nil || n.Parent.DataAtom != atom.Script {
				s = html.EscapeString(s)
			}
		}
		s = strings.TrimSpace(s)
		if s != "" {
			t.printIndent(w, level)
			fmt.Fprintln(w, s)
		}
	case html.ElementNode:
		t.printIndent(w, level)
		// TODO: allow pre with color
		if n.DataAtom == atom.Pre && !pupPrintColor && pupPreformatted {
			t.printPre(w, n)
			fmt.Fprintln(w)
			return
		}
		if pupPrintColor {
			tokenColor.Fprint(w, "<")
			tagColor.Fprintf(w, "%s", n.Data)
		} else {
			fmt.Fprintf(w, "<%s", n.Data)
		}
		for _, a := range n.Attr {
			val := a.Val
			if pupEscapeHTML {
				val = html.EscapeString(val)
			}
			if pupPrintColor {
				fmt.Fprint(w, " ")
				attrKeyColor.Fprintf(w, "%s", a.Key)
				tokenColor.Fprint(w, "=")
				quoteColor.Fprintf(w, `"%s"`, val)
			} else {
				fmt.Fprintf(w, ` %s="%s"`, a.Key, val)
			}
		}
		if pupPrintColor {
			tokenColor.Fprintln(w, ">")
		} else {
			fmt.Fprintln(w, ">")
		}
		if !isVoidElement(n) {
			t.printChildren(w, n, level+1)
			t.printIndent(w, level)
			if pupPrintColor {
				tokenColor.Fprint(w, "</")
				tagColor.Fprintf(w, "%s", n.Data)
				tokenColor.Fprintln(w, ">")
			} else {
				fmt.Fprintf(w, "</%s>\n", n.Data)
			}
		}
	case html.CommentNode:
		t.printIndent(w, level)
		data := n.Data
		if pupEscapeHTML {
			data = html.EscapeString(data)
		}
		if pupPrintColor {
			commentColor.Fprintf(w, "<!--%s-->\n", data)
		} else {
			fmt.Fprintf(w, "<!--%s-->\n", data)
		}
		t.printChildren(w, n, level)
	case html.DoctypeNode, html.DocumentNode:
		t.printChildren(w, n, level)
	}
}

func (t TreeDisplayer) printChildren(w io.Writer, n *html.Node, level int) {
	if pupMaxPrintLevel > -1 {
		if level >= pupMaxPrintLevel {
			t.printIndent(w, level)
			fmt.Fprintln(w, "...")
			return
		}
	}
	child := n.FirstChild
	for child != nil {
		t.printNode(w, child, level)
		child = child.NextSibling
	}
}

func (t TreeDisplayer) printIndent(w io.Writer, level int) {
	for ; level > 0; level-- {
		fmt.Fprint(w, pupIndentString)
	}
}

// Print the text of a node
type TextDisplayer struct{}

func (t TextDisplayer) Display(w io.Writer, nodes []*html.Node) {
	for _, node := range nodes {
		if node.Type == html.TextNode {
			data := node.Data
			if pupEscapeHTML {
				// don't escape javascript
				if node.Parent == nil || node.Parent.DataAtom != atom.Script {
					data = html.EscapeString(data)
				}
			}
			fmt.Fprintln(w, data)
		}
		children := []*html.Node{}
		child := node.FirstChild
		for child != nil {
			children = append(children, child)
			child = child.NextSibling
		}
		t.Display(w, children)
	}
}

// Print the attribute of a node
type AttrDisplayer struct {
	Attr string
}

func (a AttrDisplayer) Display(w io.Writer, nodes []*html.Node) {
	for _, node := range nodes {
		attributes := node.Attr
		for _, attr := range attributes {
			if attr.Key == a.Attr {
				val := attr.Val
				if pupEscapeHTML {
					val = html.EscapeString(val)
				}
				fmt.Fprintf(w, "%s\n", val)
			}
		}
	}
}

// Print nodes as a JSON list
type JSONDisplayer struct{}

// returns a jsonifiable struct
func jsonify(node *html.Node) map[string]interface{} {
	vals := map[string]interface{}{}
	if len(node.Attr) > 0 {
		for _, attr := range node.Attr {
			if pupEscapeHTML {
				vals[attr.Key] = html.EscapeString(attr.Val)
			} else {
				vals[attr.Key] = attr.Val
			}
		}
	}
	vals["tag"] = node.Data
	children := []interface{}{}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.ElementNode:
			children = append(children, jsonify(child))
		case html.TextNode:
			text := strings.TrimSpace(child.Data)
			if text != "" {
				if pupEscapeHTML {
					// don't escape javascript
					if node.DataAtom != atom.Script {
						text = html.EscapeString(text)
					}
				}
				// if there is already text we'll append it
				currText, ok := vals["text"]
				if ok {
					text = fmt.Sprintf("%s %s", currText, text)
				}
				vals["text"] = text
			}
		case html.CommentNode:
			comment := strings.TrimSpace(child.Data)
			if pupEscapeHTML {
				comment = html.EscapeString(comment)
			}
			currComment, ok := vals["comment"]
			if ok {
				comment = fmt.Sprintf("%s %s", currComment, comment)
			}
			vals["comment"] = comment
		}
	}
	if len(children) > 0 {
		vals["children"] = children
	}
	return vals
}

func (j JSONDisplayer) Display(w io.Writer, nodes []*html.Node) {
	var data []byte
	var err error
	jsonNodes := []map[string]interface{}{}
	for _, node := range nodes {
		jsonNodes = append(jsonNodes, jsonify(node))
	}
	data, err = json.MarshalIndent(&jsonNodes, "", pupIndentString)
	if err != nil {
		panic("Could not jsonify nodes")
	}
	fmt.Fprintf(w, "%s\n", data)
}

// Print the number of features returned
type NumDisplayer struct{}

func (d NumDisplayer) Display(w io.Writer, nodes []*html.Node) {
	fmt.Fprintln(w, len(nodes))
}

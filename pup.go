package main

import (
	"fmt"
	"os"

	"golang.org/x/net/html"
)

//      _=,_
//   o_/6 /#\
//   \__ |##/
//    ='|--\
//      /   #'-.
//      \#|_   _'-. /
//       |/ \_( # |"
//      C/ ,--___/

func main() {
	// process flags and arguments
	cmds, err := ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(2)
	}

	// Parse the input and get the root node
	root, err := ParseHTML(pupIn, pupCharset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(2)
	}
	pupIn.Close()

	if err := runSelectors(os.Stdout, cmds, root); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(2)
	}

}

func RemoveInverseMatches(root *html.Node, selectedNodes []*html.Node) {
	var remove []*html.Node
	for node := root.FirstChild; node != nil; node = node.NextSibling {
		if HtmlNodeInList(node, selectedNodes) {
			remove = append(remove, node)
		} else {
			RemoveInverseMatches(node, selectedNodes)
		}
	}
	for _, rm := range remove {
		root.RemoveChild(rm)
	}
}

func HtmlNodeInList(a *html.Node, list []*html.Node) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

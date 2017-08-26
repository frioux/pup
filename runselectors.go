package main

import (
	"fmt"
	"io"

	"golang.org/x/net/html"
)

func runSelectors(w io.Writer,  cmds []string, root *html.Node) error {
	// Parse the selectors
	selectorFuncs := []SelectorFunc{}
	funcGenerator := Select
	var cmd string
	for len(cmds) > 0 {
		cmd, cmds = cmds[0], cmds[1:]
		if len(cmds) == 0 {
			if err := ParseDisplayer(cmd); err == nil {
				continue
			}
		}
		switch cmd {
		case "*": // select all
			continue
		case ">":
			funcGenerator = SelectFromChildren
		case "+":
			funcGenerator = SelectNextSibling
		case ",": // nil will signify a comma
			selectorFuncs = append(selectorFuncs, nil)
		default:
			selector, err := ParseSelector(cmd)
			if err != nil {
				return fmt.Errorf("selector parsing error: %w", err)
			}
			selectorFuncs = append(selectorFuncs, funcGenerator(selector))
			funcGenerator = Select
		}
	}

	selectedNodes := []*html.Node{}
	currNodes := []*html.Node{root}
	for _, selectorFunc := range selectorFuncs {
		if selectorFunc == nil { // hit a comma
			selectedNodes = append(selectedNodes, currNodes...)
			currNodes = []*html.Node{root}
		} else {
			currNodes = selectorFunc(currNodes)
		}
	}
	selectedNodes = append(selectedNodes, currNodes...)
	if pupInvertSelect {
		RemoveInverseMatches(root, selectedNodes)
		pupDisplayer.Display(w, []*html.Node{root})
	} else {
		pupDisplayer.Display(w, selectedNodes)
	}

	return nil
}

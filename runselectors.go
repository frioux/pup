package main

import (
	"fmt"
	"os"

	"golang.org/x/net/html"
)

func runSelectors(cmds []string, root *html.Node) error {
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
	pupDisplayer.Display(os.Stdout, selectedNodes)

	return nil
}

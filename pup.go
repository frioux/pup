package main

import (
	"fmt"
	"os"
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

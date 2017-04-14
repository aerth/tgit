package main

import (
	"github.com/aerth/tgit/git"
	"os"
)

func main() {
	var source, destination string
	if len(os.Args) <= 2 || len(os.Args) >= 5 {
		printusage()
		os.Exit(111)
	}

	source = os.Args[2]

	if len(os.Args) == 4 {
		destination = os.Args[3]
	}

	switch os.Args[1] {
	case "clone":
		err := git.Clone(source, destination)
		if err != nil {
			println(err.Error())
			os.Exit(111)
		}
		os.Exit(0)
	default:
		println("action not supported")
		printusage()
		os.Exit(111)
	}
}

func printusage() {
	println("usage: git clone <source-repository> [<destination>]")
}

package main

import (
	"os"

	"github.com/elm-as/elmasguard/cli"
)

func main() {
	exitCode := cli.Execute(os.Args[1:])
	os.Exit(exitCode)
}


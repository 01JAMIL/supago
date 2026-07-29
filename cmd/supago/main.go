package main

import (
	"fmt"
	"os"

	"github.com/01JAMIL/supago.git/internal/cli"
)

func main() {
	root := cli.NewRootCmd()
	root.AddCommand(cli.NewInitCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

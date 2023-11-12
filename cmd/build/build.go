package main

import (
	"log"

	"github.com/lucas-ingemar/packtrak/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(cmd.GetRootCmd(), "./docs/cmd")
	if err != nil {
		log.Fatal(err)
	}
}

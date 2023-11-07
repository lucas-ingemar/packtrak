package main

import (
	"log"

	// "k8s.io/kubernetes/pkg/kubectl/cmd"
	// cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/lucas-ingemar/packtrak/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	// kubectl := cmd.NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, ioutil.Discard, ioutil.Discard)
	err := doc.GenMarkdownTree(cmd.Hej(), "./docs/cmd")
	if err != nil {
		log.Fatal(err)
	}
}

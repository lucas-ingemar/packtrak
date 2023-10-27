package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/mdnf"
	"github.com/lucas-ingemar/mdnf/internal/shared"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List status of dnf packages",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		packages, err := config.ReadPackagesConfig()
		if err != nil {
			panic(err)
		}

		state, err := config.ReadState()
		if err != nil {
			panic(err)
		}

		_, _, _, err = cmdListPackages(cmd.Context(), packages, state)
		if err != nil {
			panic(err)
		}
	},
}

func cmdListPackages(ctx context.Context, packages shared.Packages, state shared.State) (installedPkgs []string, missingPkgs []string, removedPkgs []string, err error) {
	fmt.Println("Listing DNF packages...")
	installedPkgs, missingPkgs, removedPkgs, err = mdnf.List(ctx, packages, state)
	if err != nil {
		return
	}

	for _, pkg := range installedPkgs {
		shared.PtermInstalled.Println(pkg)
	}

	for _, pkg := range missingPkgs {
		shared.PtermMissing.Println(pkg)
	}

	for _, pkg := range removedPkgs {
		shared.PtermRemoved.Println(pkg)
	}

	infoStrings := []string{}
	if len(installedPkgs) > 0 {
		infoStrings = append(infoStrings, shared.PtermInstalled.Sprintf("%d in sync", len(installedPkgs)))
	}
	if len(missingPkgs) > 0 {
		infoStrings = append(infoStrings, shared.PtermMissing.Sprintf("%d to install", len(missingPkgs)))
	}
	if len(removedPkgs) > 0 {
		infoStrings = append(infoStrings, shared.PtermRemoved.Sprintf("%d to remove", len(removedPkgs)))
	}

	if len(infoStrings) > 0 {
		fmt.Println("\n" + strings.Join(infoStrings, "   "))
	} else {
		shared.PtermGreen.Printfln("All packages up to date")
	}

	return
}

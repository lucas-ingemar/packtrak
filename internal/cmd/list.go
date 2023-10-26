package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
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
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		//FIXME: should be in an init somewhere
		packages, err := config.ReadPackagesConfig()
		if err != nil {
			panic(err)
		}

		_, _, err = cmdListPackages(cmd.Context(), packages)
		if err != nil {
			panic(err)
		}
		// fmt.Println("Listing DNF packages...")
		// installedPkgs, missingPkgs, err := mdnf.List(cmd.Context(), packages)
		// if err != nil {
		// 	panic(err)
		// }

		// for _, pkg := range installedPkgs {
		// 	color.Green(" %s", pkg)
		// }

		// for _, pkg := range missingPkgs {
		// 	color.Red(" %s", pkg)
		// }

		// fmt.Println("")
		// if len(missingPkgs) > 0 {
		// 	color.Red("%d package(s) missing", len(missingPkgs))
		// } else {
		// 	color.Green("All packages installed")
		// }
	},
}

func cmdListPackages(ctx context.Context, packages shared.Packages) (installedPkgs []string, missingPkgs []string, err error) {
	fmt.Println("Listing DNF packages...")
	installedPkgs, missingPkgs, err = mdnf.List(ctx, packages)
	if err != nil {
		return
	}

	for _, pkg := range installedPkgs {
		shared.PtermFound.Println(pkg)
	}

	for _, pkg := range missingPkgs {
		shared.PtermNotFound.Println(pkg)
	}

	fmt.Println("")
	if len(missingPkgs) > 0 {
		color.Red("%d package(s) missing", len(missingPkgs))
	} else {
		color.Green("All packages installed")
	}

	return
}

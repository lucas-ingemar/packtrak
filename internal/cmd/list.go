package cmd

import (
	"fmt"
	"strings"

	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/packagemanagers"
	"github.com/lucas-ingemar/mdnf/internal/shared"
	"github.com/spf13/cobra"
)

func initList(state shared.State) {
	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()].AddCommand(&cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List status of %s packages", pm.Name()),
			Args:  cobra.NoArgs,
			Run:   generateListCmd(pm, config.Packages[pm.Name()], state),
		})
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func generateListCmd(pm packagemanagers.PackageManager, pmPackages shared.PmPackages, state shared.State) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var err error
		pkgsSynced := map[string][]string{}
		pkgsInstall := map[string][]string{}
		pkgsRemove := map[string][]string{}
		fmt.Printf("Listing %s packages...\n", pm.Name())
		pkgsSynced[pm.Name()], pkgsInstall[pm.Name()], pkgsRemove[pm.Name()], err = pm.List(cmd.Context(), config.Packages[pm.Name()], state)
		if err != nil {
			panic(err)
		}

		printPackageList(pkgsSynced, pkgsInstall, pkgsRemove)
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List status of dnf packages",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// packages, err := config.ReadPackagesConfig()
		// if err != nil {
		// 	panic(err)
		// }

		state, err := config.ReadState()
		if err != nil {
			panic(err)
		}

		// _, _, _, err = cmdListPackages(cmd.Context(), packages, state)
		// if err != nil {
		// 	panic(err)
		// }

		pkgsSynced := map[string][]string{}
		pkgsInstall := map[string][]string{}
		pkgsRemove := map[string][]string{}
		for _, pm := range packagemanagers.PackageManagers {
			fmt.Printf("Listing %s packages...\n", pm.Name())
			pkgsSynced[pm.Name()], pkgsInstall[pm.Name()], pkgsRemove[pm.Name()], err = pm.List(cmd.Context(), config.Packages[pm.Name()], state)
			if err != nil {
				panic(err)
			}
		}

		printPackageList(pkgsSynced, pkgsInstall, pkgsRemove)

	},
}

func printPackageList(pkgsSynced, pkgsInstall, pkgsRemove map[string][]string) {
	noSync, noInstall, noRemove := 0, 0, 0
	for _, pm := range packagemanagers.PackageManagers {
		for _, pkg := range pkgsSynced[pm.Name()] {
			shared.PtermInstalled.Printfln("%s %s", pm.Icon(), pkg)
			noSync++
		}

		for _, pkg := range pkgsInstall[pm.Name()] {
			shared.PtermMissing.Printfln("%s %s", pm.Icon(), pkg)
			noInstall++
		}

		for _, pkg := range pkgsRemove[pm.Name()] {
			shared.PtermRemoved.Printfln("%s %s", pm.Icon(), pkg)
			noRemove++
		}
	}

	infoStrings := []string{}
	if noSync > 0 {
		infoStrings = append(infoStrings, shared.PtermInstalled.Sprintf("%d in sync", noSync))
	}
	if noInstall > 0 {
		infoStrings = append(infoStrings, shared.PtermMissing.Sprintf("%d to install", noInstall))
	}
	if noRemove > 0 {
		infoStrings = append(infoStrings, shared.PtermRemoved.Sprintf("%d to remove", noRemove))
	}

	if len(infoStrings) > 0 {
		fmt.Println("\n" + strings.Join(infoStrings, "   "))
	} else {
		shared.PtermGreen.Printfln("All packages up to date")
	}
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/spf13/cobra"
)

func initList() {
	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()].AddCommand(&cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List status of %s packages", pm.Name()),
			Args:  cobra.NoArgs,
			Run:   generateListCmd([]packagemanagers.PackageManager{pm}),
		})
	}

	rootCmd.AddCommand(listGlobalCmd)
}

func generateListCmd(pms []packagemanagers.PackageManager) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var err error
		tx := state.Begin()

		pkgsSynced := map[string][]string{}
		pkgsInstall := map[string][]string{}
		pkgsRemove := map[string][]string{}

		for _, pm := range pms {
			pkgsState := []string{}
			fmt.Printf("Listing %s packages...\n", pm.Name())
			pkgsSynced[pm.Name()], pkgsInstall[pm.Name()], pkgsRemove[pm.Name()], err = pm.List(cmd.Context(), tx, config.Packages[pm.Name()])
			if err != nil {
				panic(err)
			}

			pkgsState = append(pkgsState, pkgsSynced[pm.Name()]...)
			pkgsState = append(pkgsState, pkgsInstall[pm.Name()]...)

			err := state.UpdatePackageState(tx, pm.Name(), pkgsState)
			if err != nil {
				panic(err)
			}

		}

		res := tx.Commit()
		if res.Error != nil {
			panic(res.Error)
		}

		printPackageList(pkgsSynced, pkgsInstall, pkgsRemove)
	}
}

var listGlobalCmd = &cobra.Command{
	Use:   "list",
	Short: "List status of dnf packages",
	Args:  cobra.NoArgs,
	Run:   generateListCmd(packagemanagers.PackageManagers),
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

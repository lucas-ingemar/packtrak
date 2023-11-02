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

	var listGlobalCmd = &cobra.Command{
		Use:   "list",
		Short: "List status of dnf packages",
		Args:  cobra.NoArgs,
		Run:   generateListCmd(packagemanagers.PackageManagers),
	}
	rootCmd.AddCommand(listGlobalCmd)
}

func generateListCmd(pms []packagemanagers.PackageManager) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var err error
		tx := state.Begin()

		// pkgsSynced := map[string][]string{}
		// pkgsInstall := map[string][]string{}
		// pkgsRemove := map[string][]string{}
		pkgStatus := map[string]shared.PackageStatus{}

		fmt.Println(pms)
		for _, pm := range pms {
			pkgsState := []shared.Package{}
			fmt.Printf("Listing %s packages...\n", pm.Name())
			// pkgsSynced[pm.Name()], pkgsInstall[pm.Name()], pkgsRemove[pm.Name()], err = pm.List(cmd.Context(), tx, config.Packages[pm.Name()])
			pkgStatus[pm.Name()], err = pm.List(cmd.Context(), tx, config.Packages[pm.Name()])
			if err != nil {
				panic(err)
			}

			// pkgsState = append(pkgsState, pkgsSynced[pm.Name()]...)
			// pkgsState = append(pkgsState, pkgsInstall[pm.Name()]...)
			// // Must include removed pkgs as well. Otherwise the state will be messed up
			// pkgsState = append(pkgsState, pkgsRemove[pm.Name()]...)

			pkgsState = append(pkgsState, pkgStatus[pm.Name()].Synced...)
			pkgsState = append(pkgsState, pkgStatus[pm.Name()].Updated...)
			pkgsState = append(pkgsState, pkgStatus[pm.Name()].Missing...)
			pkgsState = append(pkgsState, pkgStatus[pm.Name()].Removed...)
			// pkgsState = append(pkgsState, [pm.Name()]...)
			// pkgsState = append(pkgsState, [pm.Name()]...)
			// pkgsState = append(pkgsState, [pm.Name()]...)

			err := state.UpdatePackageState(tx, pm.Name(), pkgsState)
			if err != nil {
				panic(err)
			}

		}

		res := tx.Commit()
		if res.Error != nil {
			panic(res.Error)
		}

		printPackageList(pkgStatus)
	}
}

func printPackageList(pkgStatus map[string]shared.PackageStatus) {
	noSynced, noUpdated, noMissing, noRemoved := 0, 0, 0, 0
	for _, pm := range packagemanagers.PackageManagers {
		for _, pkg := range pkgStatus[pm.Name()].Synced {
			shared.PtermInstalled.Printfln("%s %s", pm.Icon(), pkg.Name)
			noSynced++
		}

		for _, pkg := range pkgStatus[pm.Name()].Updated {
			shared.PtermInstalled.Printfln("%s %s, %s -> %s", pm.Icon(), pkg.Name, pkg.Version, pkg.LatestVersion)
			noUpdated++
		}

		for _, pkg := range pkgStatus[pm.Name()].Missing {
			shared.PtermMissing.Printfln("%s %s", pm.Icon(), pkg.Name)
			noMissing++
		}

		for _, pkg := range pkgStatus[pm.Name()].Removed {
			shared.PtermRemoved.Printfln("%s %s", pm.Icon(), pkg.Name)
			noRemoved++
		}
	}

	infoStrings := []string{}
	if noSynced > 0 {
		infoStrings = append(infoStrings, shared.PtermInstalled.Sprintf("%d in sync", noSynced))
	}
	if noMissing > 0 {
		infoStrings = append(infoStrings, shared.PtermMissing.Sprintf("%d to install", noMissing))
	}
	if noRemoved > 0 {
		infoStrings = append(infoStrings, shared.PtermRemoved.Sprintf("%d to remove", noRemoved))
	}

	if len(infoStrings) > 0 {
		fmt.Println("\n" + strings.Join(infoStrings, "   "))
	} else {
		shared.PtermGreen.Printfln("All packages up to date")
	}
}

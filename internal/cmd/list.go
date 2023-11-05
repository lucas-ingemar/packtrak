package cmd

import (
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
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
			Run:   generateListCmd([]shared.PackageManager{pm}),
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

func generateListCmd(pms []shared.PackageManager) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), pms, shared.CommandList) {
			panic("sudo access not granted")
		}

		var err error

		tx := state.Begin()

		depStatus := map[string]shared.DependenciesStatus{}
		pkgStatus := map[string]shared.PackageStatus{}

		for _, pm := range pms {
			// pkgsState := []shared.Package{}
			fmt.Printf("Listing %s dependencies...\n", pm.Name())
			// FIXME: Manifestfilter
			depStatus[pm.Name()], err = pm.ListDependencies(cmd.Context(), tx, manifest.Manifest.Pm(pm.Name()).Global.Dependencies)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Listing %s packages...\n", pm.Name())
			// FIXME: Manifestfilter
			pkgStatus[pm.Name()], err = pm.ListPackages(cmd.Context(), tx, manifest.Manifest.Pm(pm.Name()).Global.Packages)
			if err != nil {
				panic(err)
			}

			// FIXME: This is currently not used. Pretty sure I dont want to sync state on list?
			// pkgsState = append(pkgsState, pkgStatus[pm.Name()].Synced...)
			// pkgsState = append(pkgsState, pkgStatus[pm.Name()].Updated...)
			// pkgsState = append(pkgsState, pkgStatus[pm.Name()].Missing...)

		}

		res := tx.Commit()
		if res.Error != nil {
			panic(res.Error)
		}
		printPackageList(depStatus, pkgStatus)
	}
}

func printPackageList(depStatus map[string]shared.DependenciesStatus, pkgStatus map[string]shared.PackageStatus) {
	noSynced, noUpdated, noMissing, noRemoved := 0, 0, 0, 0

	fmt.Println("\nDependencies:")
	for _, pm := range packagemanagers.PackageManagers {
		for _, dep := range depStatus[pm.Name()].Synced {
			shared.PtermInstalled.Printfln("%s %s", pm.Icon(), dep.Name)
			noSynced++
		}

		for _, dep := range depStatus[pm.Name()].Missing {
			shared.PtermMissing.Printfln("%s %s", pm.Icon(), dep.Name)
			noMissing++
		}

		for _, dep := range depStatus[pm.Name()].Removed {
			shared.PtermRemoved.Printfln("%s %s", pm.Icon(), dep.Name)
			noRemoved++
		}
	}

	fmt.Println("\nPackages:")
	for _, pm := range packagemanagers.PackageManagers {
		for _, pkg := range pkgStatus[pm.Name()].Synced {
			shared.PtermInstalled.Printfln("%s %s", pm.Icon(), pkg.Name)
			noSynced++
		}

		for _, pkg := range pkgStatus[pm.Name()].Updated {
			shared.PtermUpdated.Printfln("%s %s %s -> %s", pm.Icon(), pkg.Name, pkg.Version, pkg.LatestVersion)
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
	if noUpdated > 0 {
		infoStrings = append(infoStrings, shared.PtermUpdated.Sprintf("%d to update", noUpdated))
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

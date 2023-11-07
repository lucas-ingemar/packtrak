package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
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

		tx := state.Begin()

		depStatus, pkgStatus, err := listStatus(cmd.Context(), tx, pms)
		if err != nil {
			panic(err)
		}

		res := tx.Commit()
		if res.Error != nil {
			panic(res.Error)
		}

		printPackageList(depStatus, pkgStatus)
	}
}

func listStatus(ctx context.Context, tx *gorm.DB, pms []shared.PackageManager) (map[string]shared.DependenciesStatus, map[string]shared.PackageStatus, error) {
	depStatus := map[string]shared.DependenciesStatus{}
	pkgStatus := map[string]shared.PackageStatus{}
	for _, pm := range pms {
		packages, dependencies, err := manifest.Filter(*manifest.Manifest.Pm(pm.Name()))
		if err != nil {
			return nil, nil, err
		}
		fmt.Printf("Listing %s dependencies...\n", pm.Name())
		depStatus[pm.Name()], err = pm.ListDependencies(ctx, tx, dependencies)
		if err != nil {
			return nil, nil, err
		}
		fmt.Printf("Listing %s packages...\n", pm.Name())
		pkgStatus[pm.Name()], err = pm.ListPackages(ctx, tx, packages)
		if err != nil {
			return nil, nil, err
		}
	}
	return depStatus, pkgStatus, nil
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

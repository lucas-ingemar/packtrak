package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync DNF to match mDNF",
	Args:  cobra.NoArgs,
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmdSync(cmd.Context())
		if err != nil {
			panic(err)
		}
	},
}

func cmdSync(ctx context.Context) (err error) {
	if !shared.MustDoSudo(ctx, packagemanagers.PackageManagers, shared.CommandSync) {
		return errors.New("sudo access not granted")
	}

	tx := state.Begin()
	defer tx.Rollback()

	// var fpkgM, fpkgU, fpkgR []shared.Package
	totUpdatedPkgs := []shared.Package{}
	totUpdatedDeps := []shared.Dependency{}

	pkgsState := map[string][]shared.Package{}
	pkgStatus := map[string]shared.PackageStatus{}

	depsState := map[string][]shared.Dependency{}
	depStatus := map[string]shared.DependenciesStatus{}

	for _, pm := range packagemanagers.PackageManagers {
		fmt.Printf("Listing %s dependencies...\n", pm.Name())
		depStatus[pm.Name()], err = pm.ListDependencies(ctx, tx, config.Packages[pm.Name()])
		if err != nil {
			panic(err)
		}
		fmt.Printf("Listing %s packages...\n", pm.Name())
		pkgStatus[pm.Name()], err = pm.ListPackages(ctx, tx, config.Packages[pm.Name()])
		if err != nil {
			return
		}
		totUpdatedDeps = append(totUpdatedDeps, depStatus[pm.Name()].Missing...)
		totUpdatedDeps = append(totUpdatedDeps, depStatus[pm.Name()].Updated...)
		totUpdatedDeps = append(totUpdatedDeps, depStatus[pm.Name()].Removed...)

		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Synced...)
		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Updated...)
		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Missing...)

		totUpdatedPkgs = append(totUpdatedPkgs, pkgStatus[pm.Name()].Missing...)
		totUpdatedPkgs = append(totUpdatedPkgs, pkgStatus[pm.Name()].Updated...)
		totUpdatedPkgs = append(totUpdatedPkgs, pkgStatus[pm.Name()].Removed...)

		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Synced...)
		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Updated...)
		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Missing...)
		// pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Removed...)
	}

	printPackageList(depStatus, pkgStatus)

	if len(totUpdatedDeps) == 0 && len(totUpdatedPkgs) == 0 {
		for _, pm := range packagemanagers.PackageManagers {
			err := state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}

			err = state.UpdateDependencyState(tx, pm.Name(), depsState[pm.Name()])
			if err != nil {
				return err
			}
		}
		return tx.Commit().Error
	}

	fmt.Println("")
	result, _ := pterm.InteractiveContinuePrinter{
		DefaultValueIndex: 0,
		DefaultText:       "Unsynced changes found in config. Do you want to sync?",
		TextStyle:         &pterm.ThemeDefault.PrimaryStyle,
		Options:           []string{"y", "n"},
		OptionsStyle:      &pterm.ThemeDefault.SuccessMessageStyle,
		SuffixStyle:       &pterm.ThemeDefault.SecondaryStyle,
		Delimiter:         ": ",
	}.Show()

	if result == "y" {
		for _, pm := range packagemanagers.PackageManagers {
			uw, err := pm.SyncDependencies(ctx, depStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = state.UpdateDependencyState(tx, pm.Name(), depsState[pm.Name()])
			if err != nil {
				return err
			}

			uw, err = pm.SyncPackages(ctx, pkgStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit().Error
}

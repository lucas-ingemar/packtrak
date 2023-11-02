package cmd

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
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
		// packages, err := config.ReadPackagesConfig()
		// if err != nil {
		// 	panic(err)
		// }

		// state, err := config.ReadState()
		// if err != nil {
		// 	panic(err)
		// }

		err := cmdSync(cmd.Context())
		if err != nil {
			panic(err)
		}
	},
}

func cmdSync(ctx context.Context) (err error) {
	tx := state.Begin()
	defer tx.Rollback()

	var fpkgI, fpkgR []string
	pkgsSynced := map[string][]string{}
	pkgsInstall := map[string][]string{}
	pkgsRemove := map[string][]string{}
	pkgsState := map[string][]string{}

	for _, pm := range packagemanagers.PackageManagers {
		fmt.Printf("Listing %s packages...\n", pm.Name())
		pkgsSynced[pm.Name()], pkgsInstall[pm.Name()], pkgsRemove[pm.Name()], err = pm.List(ctx, tx, config.Packages[pm.Name()])
		if err != nil {
			return
		}
		fpkgI = append(fpkgI, pkgsInstall[pm.Name()]...)
		fpkgR = append(fpkgR, pkgsRemove[pm.Name()]...)

		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgsSynced[pm.Name()]...)
		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgsInstall[pm.Name()]...)
	}

	printPackageList(pkgsSynced, pkgsInstall, pkgsRemove)

	if len(fpkgI) == 0 && len(fpkgR) == 0 {
		//FIXME: Must update state here aswell. It will update when package exists
		for _, pm := range packagemanagers.PackageManagers {
			//FIXME: This have to be enabled somehow
			// if config.DnfEnabled {
			err := state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
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
			//FIXME: This have to be enabled somehow
			// if config.DnfEnabled {
			uw, err := pm.Sync(ctx, pkgsInstall[pm.Name()], pkgsRemove[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}
			// }
		}
	}

	return tx.Commit().Error
}

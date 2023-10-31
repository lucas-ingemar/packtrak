package cmd

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/packagemanagers"
	"github.com/lucas-ingemar/mdnf/internal/shared"
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

		state, err := config.ReadState()
		if err != nil {
			panic(err)
		}

		err = cmdSync(cmd.Context(), state)
		if err != nil {
			panic(err)
		}
	},
}

func cmdSync(ctx context.Context, state shared.State) (err error) {
	var fpkgI, fpkgR []string
	pkgsSynced := map[string][]string{}
	pkgsInstall := map[string][]string{}
	pkgsRemove := map[string][]string{}

	for _, pm := range packagemanagers.PackageManagers {
		fmt.Printf("Listing %s packages...\n", pm.Name())
		pkgsSynced[pm.Name()], pkgsInstall[pm.Name()], pkgsRemove[pm.Name()], err = pm.List(ctx, config.Packages[pm.Name()], state)
		if err != nil {
			return
		}
		fpkgI = append(fpkgI, pkgsInstall[pm.Name()]...)
		fpkgR = append(fpkgR, pkgsRemove[pm.Name()]...)
	}

	printPackageList(pkgsSynced, pkgsInstall, pkgsRemove)

	if len(fpkgI) == 0 && len(fpkgR) == 0 {
		return config.NewState(config.Packages)
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
			// }
		}
	}

	return config.NewState(config.Packages)
}

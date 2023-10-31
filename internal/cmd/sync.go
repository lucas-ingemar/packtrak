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
		packages, err := config.ReadPackagesConfig()
		if err != nil {
			panic(err)
		}

		state, err := config.ReadState()
		if err != nil {
			panic(err)
		}

		err = cmdSync(cmd.Context(), packages, state)
		if err != nil {
			panic(err)
		}
	},
}

func cmdSync(ctx context.Context, packages shared.Packages, state shared.State) error {
	_, missingPkgs, removedPkgs, err := cmdListPackages(ctx, packages, state)
	if err != nil {
		return err
	}

	if len(missingPkgs) == 0 && len(removedPkgs) == 0 {
		return config.NewState(packages)
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
		if config.DnfEnabled {
			uw, err := packagemanagers.PackageManagers[0].Sync(ctx, missingPkgs, removedPkgs)
			_ = uw
			if err != nil {
				return err
			}
		}
	}

	return config.NewState(packages)
}

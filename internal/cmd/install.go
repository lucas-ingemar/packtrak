package cmd

import (
	"fmt"

	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/dnf"
	"github.com/lucas-ingemar/mdnf/internal/mdnf"
	"github.com/lucas-ingemar/mdnf/internal/shared"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install a package or packages on your system",
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		pkgs, err := dnf.ListAvailable(cmd.Context(), toComplete)
		if err != nil {
			pkgs = []string{}
		}
		return pkgs, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		args = lo.Uniq(args)

		cPackages, err := config.ReadPackagesConfig()
		if err != nil {
			panic(err)
		}

		pkgsToAdd := []string{}
		for _, arg := range args {
			if lo.Contains(cPackages.Global.Packages, arg) {
				// color.Yellow(" '%s' is already present in packages file", arg)
				shared.PtermWarning.Printfln("'%s' is already present in packages file", arg)
				continue
			}
			pkgsToAdd = append(pkgsToAdd, arg)
		}

		cPackages, err = mdnf.AddPackages(cPackages, pkgsToAdd)
		if err != nil {
			panic(err)
		}

		installedPkgs, missingPkgs, err := cmdListPackages(cmd.Context(), cPackages)
		if err != nil {
			panic(err)
		}
		_ = installedPkgs
		_ = missingPkgs

		result, _ := pterm.InteractiveContinuePrinter{
			DefaultValueIndex: 0,
			DefaultText:       fmt.Sprintf("Do you want to %s %d packages?", "hejsan", 3),
			TextStyle:         &pterm.ThemeDefault.PrimaryStyle,
			Options:           []string{"y", "n"},
			OptionsStyle:      &pterm.ThemeDefault.SuccessMessageStyle,
			SuffixStyle:       &pterm.ThemeDefault.SecondaryStyle,
			Delimiter:         ": ",
		}.Show()
		fmt.Println(result)

		// pterm.Warning.Printfln("One or more packages failed to %s", "hej")

		// dnf.ListAvailable(cmd.Context(), "terr")
		// // //FIXME: should be in an init somewhere
		// // packages, err := config.ReadPackagesConfig()
		// // if err != nil {
		// // 	panic(err)
		// // }

		// // fmt.Println("Listing DNF packages...")
		// // installedPkgs, missingPkgs, err := mdnf.List(packages)
		// // if err != nil {
		// // 	panic(err)
		// // }

		// // for _, pkg := range installedPkgs {
		// // 	color.Green(" %s", pkg)
		// // }

		// // for _, pkg := range missingPkgs {
		// // 	color.Red(" %s", pkg)
		// // }

		// // fmt.Println("")
		// // if len(missingPkgs) > 0 {
		// // 	color.Red("%d package(s) missing", len(missingPkgs))
		// // } else {
		// // 	color.Green("All packages installed")
		// // }
		// fmt.Println("hej")
	},
}

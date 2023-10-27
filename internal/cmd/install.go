package cmd

import (
	"fmt"

	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/dnf"
	"github.com/lucas-ingemar/mdnf/internal/mdnf"
	"github.com/lucas-ingemar/mdnf/internal/shared"
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

		state, err := config.ReadState()
		if err != nil {
			panic(err)
		}

		pkgsToAdd := []string{}
		warningPrinted := false
		for _, arg := range args {
			if lo.Contains(cPackages.Dnf.Global.Packages, arg) {
				shared.PtermWarning.Printfln("'%s' is already present in packages file", arg)
				warningPrinted = true
				continue
			}
			pkgsToAdd = append(pkgsToAdd, arg)
		}
		if warningPrinted {
			fmt.Println("")
		}

		cPackages, err = mdnf.AddPackages(cPackages, pkgsToAdd)
		if err != nil {
			panic(err)
		}

		err = cmdSync(cmd.Context(), cPackages, state)
		if err != nil {
			panic(err)
		}

		err = config.SavePackages(cPackages)
		if err != nil {
			panic(err)
		}
	},
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/mdnf"
	"github.com/lucas-ingemar/mdnf/internal/shared"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a package or packages on your system",
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		packages, err := config.ReadPackagesConfig()
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		return lo.Filter(packages.Dnf.Global.Packages,
				func(item string, index int) bool {
					return strings.HasPrefix(item, toComplete)
				}),
			cobra.ShellCompDirectiveNoFileComp
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

		pkgsToRemove := []string{}
		warningPrinted := false
		for _, arg := range args {
			if !lo.Contains(cPackages.Dnf.Global.Packages, arg) {
				shared.PtermWarning.Printfln("'%s' is not present in packages file", arg)
				warningPrinted = true
				continue
			}
			pkgsToRemove = append(pkgsToRemove, arg)
		}
		if warningPrinted {
			fmt.Println("")
		}

		cPackages, err = mdnf.RemovePackages(cPackages, pkgsToRemove)
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

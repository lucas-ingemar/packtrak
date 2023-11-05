package cmd

import (
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func initRemove() {
	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()].AddCommand(&cobra.Command{
			Use:               "remove",
			Short:             "remove a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateRemoveValidArgsFunc(pm, config.Packages[pm.Name()]),
			Run:               generateRemoveCmd(pm, config.Packages[pm.Name()]),
		})
	}
}

func generateRemoveValidArgsFunc(pm shared.PackageManager, pmPackages shared.PmPackages) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return lo.Filter(pm.GetPackageNames(cmd.Context(), pmPackages),
				func(item string, index int) bool {
					return strings.HasPrefix(item, toComplete)
				}),
			cobra.ShellCompDirectiveNoFileComp
	}
}

func generateRemoveCmd(pm shared.PackageManager, pmPackages shared.PmPackages) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), []shared.PackageManager{pm}, shared.CommandRemove) {
			panic("sudo access not granted")
		}
		args = lo.Uniq(args)

		pkgsToRemove := []string{}
		warningPrinted := false
		for _, arg := range args {
			if !lo.Contains(pm.GetPackageNames(cmd.Context(), pmPackages), arg) {
				shared.PtermWarning.Printfln("'%s' is not present in packages file", arg)
				warningPrinted = true
				continue
			}
			pkgsToRemove = append(pkgsToRemove, arg)
		}

		pmPackages, userWarnings, err := pm.Remove(cmd.Context(), pmPackages, pkgsToRemove)
		if err != nil {
			panic(err)
		}

		for _, uw := range userWarnings {
			shared.PtermWarning.Println(uw)
			warningPrinted = true
		}

		if warningPrinted {
			fmt.Println("")
		}

		config.Packages[pm.Name()] = pmPackages

		err = cmdSync(cmd.Context())
		if err != nil {
			panic(err)
		}

		err = config.SavePackages()
		if err != nil {
			panic(err)
		}
	}
}

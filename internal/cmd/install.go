package cmd

import (
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func initInstall() {
	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()].AddCommand(&cobra.Command{
			Use:   "install",
			Short: "install a package or packages on your system",
			Args:  cobra.MinimumNArgs(1),
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				pkgs, err := pm.InstallValidArgs(cmd.Context(), toComplete)
				if err != nil {
					pkgs = []string{}
				}
				return pkgs, cobra.ShellCompDirectiveNoFileComp
			},
			Run: generateInstallCmd(pm, config.Packages[pm.Name()]),
		})
	}
}

func generateInstallCmd(pm packagemanagers.PackageManager, pmPackages shared.PmPackages) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		args = lo.Uniq(args)

		pkgsToAdd := []string{}
		warningPrinted := false
		for _, arg := range args {
			if lo.Contains(pmPackages.Global.Packages, arg) {
				shared.PtermWarning.Printfln("'%s' is already present in packages file", arg)
				warningPrinted = true
				continue
			}
			pkgsToAdd = append(pkgsToAdd, arg)
		}

		pmPackages, userWarnings, err := pm.Add(cmd.Context(), pmPackages, pkgsToAdd)
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

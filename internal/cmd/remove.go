package cmd

import (
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
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
			ValidArgsFunction: generateRemoveValidArgsFunc(pm, manifest.Manifest.Pm(pm.Name())),
			Run:               generateRemoveCmd(pm, manifest.Manifest.Pm(pm.Name())),
		})
	}
}

func generateRemoveValidArgsFunc(pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// FIXME: Manifestfilter
		return lo.Filter(pm.GetPackageNames(cmd.Context(), pmManifest.Global.Packages),
				func(item string, index int) bool {
					return strings.HasPrefix(item, toComplete)
				}),
			cobra.ShellCompDirectiveNoFileComp
	}
}

func generateRemoveCmd(pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), []shared.PackageManager{pm}, shared.CommandRemove) {
			panic("sudo access not granted")
		}
		args = lo.Uniq(args)

		pkgsToRemove := []string{}
		warningPrinted := false
		for _, arg := range args {
			// FIXME: Manifestfilter
			if !lo.Contains(pm.GetPackageNames(cmd.Context(), pmManifest.Global.Packages), arg) {
				shared.PtermWarning.Printfln("'%s' is not present in packages file", arg)
				warningPrinted = true
				continue
			}
			pkgsToRemove = append(pkgsToRemove, arg)
		}

		//FIXME: pkgsToRemove might be enough!? concat arrays here instead...
		pmPackages, userWarnings, err := pm.Remove(cmd.Context(), pmManifest.Global.Packages, pkgsToRemove)
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

		// config.Packages[pm.Name()] = pmPackages
		pmManifest.Global.RemovePackages(pmPackages)

		err = cmdSync(cmd.Context())
		if err != nil {
			panic(err)
		}

		err = manifest.SaveManifest(config.ManifestFile)
		if err != nil {
			panic(err)
		}
	}
}

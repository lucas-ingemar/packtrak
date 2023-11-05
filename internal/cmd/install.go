package cmd

import (
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func initInstall() {
	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()].AddCommand(&cobra.Command{
			Use:               "install",
			Short:             "install a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateInstallValidArgsFunc(pm, manifest.Manifest.Pm(pm.Name())),
			Run:               generateInstallCmd(pm, manifest.Manifest.Pm(pm.Name())),
		})
	}
}

func generateInstallValidArgsFunc(pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		pkgs, err := pm.InstallValidArgs(cmd.Context(), toComplete)
		if err != nil {
			pkgs = []string{}
		}
		return pkgs, cobra.ShellCompDirectiveNoFileComp
	}
}

func generateInstallCmd(pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), []shared.PackageManager{pm}, shared.CommandInstall) {
			panic("sudo access not granted")
		}

		args = lo.Uniq(args)
		pkgsToAdd := []string{}
		warningPrinted := false
		for _, arg := range args {
			// FIXME: Manifestfilter
			if lo.Contains(pmManifest.Global.Packages, arg) {
				shared.PtermWarning.Printfln("'%s' is already present in packages file", arg)
				warningPrinted = true
				continue
			}
			pkgsToAdd = append(pkgsToAdd, arg)
		}

		//FIXME: pkgsToAdd might be enough!? concat arrays here instead...
		// FIXME: Manifestfilter
		pmPackages, userWarnings, err := pm.Add(cmd.Context(), pmManifest.Global.Packages, pkgsToAdd)
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

		// FIXME: Manifestfilter
		manifest.Manifest.Pm(pm.Name()).Global.AddPackages(pmPackages)
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

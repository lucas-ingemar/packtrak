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
		installCmd := &cobra.Command{
			Use:               "install",
			Short:             "install a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateInstallValidArgsFunc(pm, manifest.Manifest.Pm(pm.Name())),
			Run:               generateInstallCmd(pm, manifest.Manifest.Pm(pm.Name())),
		}
		installCmd.PersistentFlags().BoolP("dependency", "d", false, "Install dependency")
		PmCmds[pm.Name()].AddCommand(installCmd)
	}
}

func generateInstallValidArgsFunc(pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		installDependency := cmd.Flag("dependency").Value.String() == "true"
		pkgs, err := pm.InstallValidArgs(cmd.Context(), toComplete, installDependency)
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
		// FIXME: Manifestfilter: Must add a conditional flag

		installDependency := cmd.Flag("dependency").Value.String() == "true"

		args = lo.Uniq(args)
		objsToAdd := []string{}
		warningPrinted := false
		for _, arg := range args {
			var objs []string
			pkgs, deps, err := manifest.Filter(*pmManifest)
			if err != nil {
				panic(err)
			}
			if installDependency {
				objs = deps
			} else {
				objs = pkgs
			}
			if lo.Contains(objs, arg) {
				shared.PtermWarning.Printfln("'%s' is already present in manifest", arg)
				warningPrinted = true
				continue
			}
			objsToAdd = append(objsToAdd, arg)
		}

		var toAdd, userWarnings []string
		var err error

		if installDependency {
			toAdd, userWarnings, err = pm.AddDependencies(cmd.Context(), objsToAdd)
			if err != nil {
				panic(err)
			}
		} else {
			toAdd, userWarnings, err = pm.AddPackages(cmd.Context(), objsToAdd)
			if err != nil {
				panic(err)
			}
		}

		for _, uw := range userWarnings {
			shared.PtermWarning.Println(uw)
			warningPrinted = true
		}

		if warningPrinted {
			fmt.Println("")
		}

		// FIXME: Manifestfilter: Must add a conditional flag
		if installDependency {
			manifest.Manifest.Pm(pm.Name()).Global.AddDependencies(toAdd)
		} else {
			manifest.Manifest.Pm(pm.Name()).Global.AddPackages(toAdd)
		}

		err = cmdSync(cmd.Context(), []shared.PackageManager{pm})
		if err != nil {
			panic(err)
		}

		err = manifest.SaveManifest(config.ManifestFile)
		if err != nil {
			panic(err)
		}
	}
}

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
		removeCmd := &cobra.Command{
			Use:               "remove",
			Short:             "remove a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateRemoveValidArgsFunc(pm, manifest.Manifest.Pm(pm.Name())),
			Run:               generateRemoveCmd(pm, manifest.Manifest.Pm(pm.Name())),
		}
		removeCmd.PersistentFlags().BoolP("dependency", "d", false, "Remove dependency")
		PmCmds[pm.Name()].AddCommand(removeCmd)
	}
}

func generateRemoveValidArgsFunc(pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// FIXME: Manifestfilter
		// FIXME: Support dependencies
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
		removeDependency := cmd.Flag("dependency").Value.String() == "true"

		args = lo.Uniq(args)

		objsToRemove := []string{}
		warningPrinted := false

		var objs []string
		pkgs, deps, err := manifest.Filter(*pmManifest)
		if err != nil {
			panic(err)
		}

		if removeDependency {
			objs = pm.GetDependencyNames(cmd.Context(), deps)
		} else {
			objs = pm.GetPackageNames(cmd.Context(), pkgs)
		}

		for _, arg := range args {
			if !lo.Contains(objs, arg) {
				shared.PtermWarning.Printfln("'%s' is not present in manifest", arg)
				warningPrinted = true
				continue
			}
			objsToRemove = append(objsToRemove, arg)
		}

		var toRemove, userWarnings []string

		if removeDependency {
			toRemove, userWarnings, err = pm.RemoveDependencies(cmd.Context(), deps, objsToRemove)
			if err != nil {
				panic(err)
			}
		} else {
			toRemove, userWarnings, err = pm.RemovePackages(cmd.Context(), pkgs, objsToRemove)
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
		// FIXME: Also needs to do something smart. Otherwise a specific conditional needs to specified to be able to remove
		if removeDependency {
			pmManifest.Global.RemoveDependencies(toRemove)
		} else {
			pmManifest.Global.RemovePackages(toRemove)
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

package cmd

import (
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func initRemove(a app.AppFace) {
	for _, pm := range packagemanagers.PackageManagers {
		removeCmd := &cobra.Command{
			Use:               "remove",
			Short:             "remove a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateRemoveValidArgsFunc(pm, manifest.Manifest.Pm(pm.Name())),
			Run:               generateRemoveCmd(a, pm, manifest.Manifest.Pm(pm.Name())),
		}
		removeCmd.PersistentFlags().BoolP("dependency", "d", false, "Remove dependency")
		PmCmds[pm.Name()].AddCommand(removeCmd)
	}
}

func generateRemoveValidArgsFunc(pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		removeDependency := cmd.Flag("dependency").Value.String() == "true"
		pkgs, deps, err := manifest.Filter(*pmManifest)
		if err != nil {
			panic(err)
		}

		if removeDependency {
			return lo.Filter(pm.GetDependencyNames(cmd.Context(), deps),
					func(item string, index int) bool {
						return strings.HasPrefix(item, toComplete)
					}),
				cobra.ShellCompDirectiveNoFileComp
		} else {
			return lo.Filter(pm.GetPackageNames(cmd.Context(), pkgs),
					func(item string, index int) bool {
						return strings.HasPrefix(item, toComplete)
					}),
				cobra.ShellCompDirectiveNoFileComp
		}
	}
}

func generateRemoveCmd(a app.AppFace, pm shared.PackageManager, pmManifest *shared.PmManifest) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), []shared.PackageManager{pm}, shared.CommandRemove) {
			panic("sudo access not granted")
		}
		removeDependency := cmd.Flag("dependency").Value.String() == "true"
		if err := a.Remove(cmd.Context(), args, pm, pmManifest, removeDependency); err != nil {
			panic(err)
		}
	}
}

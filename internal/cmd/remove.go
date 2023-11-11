package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/cobra"
)

func initRemove(a app.AppFace) {
	for _, managerName := range a.ListManagers() {
		removeCmd := &cobra.Command{
			Use:               "remove",
			Short:             "remove a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateRemoveValidArgsFunc(a, managerName),
			Run:               generateRemoveCmd(a, managerName),
		}
		removeCmd.PersistentFlags().BoolP("dependency", "d", false, "Remove dependency")
		PmCmds[managerName].AddCommand(removeCmd)
	}
}

func generateRemoveValidArgsFunc(a app.AppFace, managerName shared.ManagerName) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var mType manifest.ManifestObjectType
		if cmd.Flag("dependency").Value.String() == "true" {
			mType = manifest.TypeDependency
		} else {
			mType = manifest.TypePackage
		}

		retArgs, err := a.RemoveValidArgsFunc(cmd.Context(), toComplete, managerName, mType)
		if err != nil {
			panic(err)
		}

		return retArgs, cobra.ShellCompDirectiveNoFileComp

		// pmManifest := a.Manifest.Pm()
		// pkgs, deps, err := manifest.Filter(*pmManifest)
		// if err != nil {
		// 	panic(err)
		// }

		// if removeDependency {
		// 	return lo.Filter(pm.GetDependencyNames(cmd.Context(), deps),
		// 			func(item string, index int) bool {
		// 				return strings.HasPrefix(item, toComplete)
		// 			}),
		// 		cobra.ShellCompDirectiveNoFileComp
		// } else {
		// 	return lo.Filter(pm.GetPackageNames(cmd.Context(), pkgs),
		// 			func(item string, index int) bool {
		// 				return strings.HasPrefix(item, toComplete)
		// 			}),
		// 		cobra.ShellCompDirectiveNoFileComp
		// }
	}
}

func generateRemoveCmd(a app.AppFace, managerName shared.ManagerName) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		// if !shared.MustDoSudo(cmd.Context(), []shared.PackageManager{pm}, shared.CommandRemove) {
		// 	panic("sudo access not granted")
		// }

		var mType manifest.ManifestObjectType
		if cmd.Flag("dependency").Value.String() == "true" {
			mType = manifest.TypeDependency
		} else {
			mType = manifest.TypePackage
		}

		// removeDependency := cmd.Flag("dependency").Value.String() == "true"
		if err := a.Remove(cmd.Context(), args, managerName, mType); err != nil {
			panic(err)
		}
	}
}

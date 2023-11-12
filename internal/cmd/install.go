package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/cobra"
)

func initInstall(a app.AppFace) {
	for _, m := range a.ListManagers() {
		installCmd := &cobra.Command{
			Use:               "install",
			Short:             "install a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateInstallValidArgsFunc(a, m),
			Run:               generateInstallCmd(a, m),
		}
		installCmd.PersistentFlags().BoolP("dependency", "d", false, "Install dependency")
		installCmd.PersistentFlags().Bool("host", false, "Install only for the current host")
		installCmd.PersistentFlags().String("group", "", "Install only for specified group")
		PmCmds[m].AddCommand(installCmd)
	}
}

func generateInstallValidArgsFunc(a app.AppFace, managerName shared.ManagerName) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		var mType manifest.ManifestObjectType
		if cmd.Flag("dependency").Value.String() == "true" {
			mType = manifest.TypeDependency
		} else {
			mType = manifest.TypePackage
		}

		pkgs, err := a.InstallValidArgsFunc(cmd.Context(), managerName, toComplete, mType)
		if err != nil {
			pkgs = []string{}
		}
		return pkgs, cobra.ShellCompDirectiveNoFileComp
	}
}

func generateInstallCmd(a app.AppFace, managerName shared.ManagerName) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var mType manifest.ManifestObjectType
		if cmd.Flag("dependency").Value.String() == "true" {
			mType = manifest.TypeDependency
		} else {
			mType = manifest.TypePackage
		}

		group := cmd.Flag("group").Value.String()
		host := cmd.Flag("host").Value.String() == "true"

		if err := a.Install(cmd.Context(), args, managerName, mType, host, group); err != nil {
			panic(err)
		}
	}
}

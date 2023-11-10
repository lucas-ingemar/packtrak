package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/cobra"
)

func initInstall(a app.AppFace) {
	for _, pm := range packagemanagers.PackageManagers {
		installCmd := &cobra.Command{
			Use:               "install",
			Short:             "install a package or packages on your system",
			Args:              cobra.MinimumNArgs(1),
			ValidArgsFunction: generateInstallValidArgsFunc(pm),
			Run:               generateInstallCmd(a, pm),
		}
		installCmd.PersistentFlags().BoolP("dependency", "d", false, "Install dependency")
		installCmd.PersistentFlags().Bool("host", false, "Install only for the current host")
		installCmd.PersistentFlags().String("group", "", "Install only for specified group")
		PmCmds[pm.Name()].AddCommand(installCmd)
	}
}

func generateInstallValidArgsFunc(pm shared.PackageManager) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		installDependency := cmd.Flag("dependency").Value.String() == "true"
		pkgs, err := pm.InstallValidArgs(cmd.Context(), toComplete, installDependency)
		if err != nil {
			pkgs = []string{}
		}
		return pkgs, cobra.ShellCompDirectiveNoFileComp
	}
}

func generateInstallCmd(a app.AppFace, pm shared.PackageManager) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), []shared.PackageManager{pm}, shared.CommandInstall) {
			panic("sudo access not granted")
		}

		var mType manifest.ManifestObjectType
		if cmd.Flag("dependency").Value.String() == "true" {
			mType = manifest.TypeDependency
		} else {
			mType = manifest.TypePackage
		}

		group := cmd.Flag("group").Value.String()
		host := cmd.Flag("host").Value.String() == "true"

		if err := a.Install(cmd.Context(), args, pm, mType, host, group); err != nil {
			panic(err)
		}
	}
}

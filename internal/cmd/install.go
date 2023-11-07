package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/machinery"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
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
		installCmd.PersistentFlags().Bool("host", false, "Install only for the current host")
		installCmd.PersistentFlags().String("group", "", "Install only for specified group")
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

		installDependency := cmd.Flag("dependency").Value.String() == "true"
		group := cmd.Flag("group").Value.String()
		host := cmd.Flag("host").Value.String() == "true"

		if err := machinery.Install(cmd.Context(), args, pm, pmManifest, installDependency, host, group); err != nil {
			panic(err)
		}
	}
}

package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/machinery"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync DNF to match mDNF",
	Args:  cobra.NoArgs,
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), packagemanagers.PackageManagers, shared.CommandSync) {
			panic("sudo access not granted")
		}
		err := machinery.Sync(cmd.Context(), packagemanagers.PackageManagers)
		if err != nil {
			panic(err)
		}
	},
}

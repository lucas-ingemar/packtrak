package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/cobra"
)

func initSync(a app.AppFace) {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync DNF to match mDNF",
		Args:  cobra.NoArgs,
		// Long:  `All software has versions. This is Hugo's`,
		Run: func(cmd *cobra.Command, args []string) {
			if !shared.MustDoSudo(cmd.Context(), packagemanagers.PackageManagers, shared.CommandSync) {
				panic("sudo access not granted")
			}
			err := a.Sync(cmd.Context(), packagemanagers.PackageManagers)
			if err != nil {
				panic(err)
			}
		},
	}
	rootCmd.AddCommand(syncCmd)
}

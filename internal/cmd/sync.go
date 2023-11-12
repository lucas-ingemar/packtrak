package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/spf13/cobra"
)

func initSync(a app.AppFace) {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync DNF to match mDNF",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := a.Sync(cmd.Context(), a.ListManagers())
			if err != nil {
				panic(err)
			}
		},
	}
	rootCmd.AddCommand(syncCmd)
}

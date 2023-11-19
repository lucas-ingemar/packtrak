package cmd

import (
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func initSync(a app.AppFace) {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync managers to match manifest",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			err := a.Sync(cmd.Context(), a.ListManagers())
			if err != nil {
				log.Fatal().Err(err).Msg("initSync")
			}
		},
	}
	rootCmd.AddCommand(syncCmd)
}

package cmd

import (
	"fmt"
	"runtime"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display app version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("packtrack (%s) version %s, %s\n", config.RepoUrl, config.Version, runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

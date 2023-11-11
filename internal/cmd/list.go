package cmd

import (
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/cobra"
)

func initList(a app.AppFace) {
	for _, m := range a.ListManagers() {
		PmCmds[m].AddCommand(&cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List status of %s packages", m),
			Args:  cobra.NoArgs,
			Run:   generateListCmd(a, []shared.ManagerName{m}),
		})
	}

	var listGlobalCmd = &cobra.Command{
		Use:   "list",
		Short: "List status of dnf packages",
		Args:  cobra.NoArgs,
		Run:   generateListCmd(a, a.ListManagers()),
	}
	rootCmd.AddCommand(listGlobalCmd)
}

func generateListCmd(a app.AppFace, pms []shared.ManagerName) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		status, err := a.ListStatus(cmd.Context(), pms)
		if err != nil {
			panic(err)
		}

		a.PrintPackageList(status)
	}
}

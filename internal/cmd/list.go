package cmd

import (
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/core"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/cobra"
)

func initList(a app.AppFace) {
	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()].AddCommand(&cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List status of %s packages", pm.Name()),
			Args:  cobra.NoArgs,
			Run:   generateListCmd(a, []shared.PackageManager{pm}),
		})
	}

	var listGlobalCmd = &cobra.Command{
		Use:   "list",
		Short: "List status of dnf packages",
		Args:  cobra.NoArgs,
		Run:   generateListCmd(a, packagemanagers.PackageManagers),
	}
	rootCmd.AddCommand(listGlobalCmd)
}

func generateListCmd(a app.AppFace, pms []shared.PackageManager) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), pms, shared.CommandList) {
			panic("sudo access not granted")
		}

		//FIXME:  begin should be inside liststatus
		// tx := state.Begin()

		depStatus, pkgStatus, err := a.ListStatus(cmd.Context(), pms)
		if err != nil {
			panic(err)
		}

		// res := tx.Commit()
		// if res.Error != nil {
		// 	panic(res.Error)
		// }

		core.PrintPackageList(depStatus, pkgStatus)
	}
}

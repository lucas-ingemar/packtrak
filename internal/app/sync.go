package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/core"
	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/pterm/pterm"
)

func (a App) Sync(ctx context.Context, managerNames []managers.ManagerName) (err error) {
	ms, error := a.Managers.GetManagers(managerNames)
	if error != nil {
		return error
	}

	if !shared.MustDoSudo(ctx, ms, shared.CommandSync) {
		panic("sudo access not granted")
	}

	depStatus, pkgStatus, err := a.ListStatus(ctx, managerNames)
	if err != nil {
		return err
	}

	pkgsState := core.UpdatedPackageState(ms, pkgStatus)
	depsState := core.UpdatedDependencyState(ms, depStatus)

	a.PrintPackageList(depStatus, pkgStatus)

	if core.CountUpdatedPkgs(ms, pkgStatus) == 0 && core.CountUpdatedDeps(ms, depStatus) == 0 {
		tx := a.State.Begin(ctx)
		defer tx.Rollback()
		for _, manager := range ms {
			err := tx.UpdatePackageState(ctx, manager.Name(), pkgsState[manager.Name()])
			if err != nil {
				return err
			}

			err = tx.UpdateDependencyState(ctx, manager.Name(), depsState[manager.Name()])
			if err != nil {
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return state.Rotate(config.StateRotations)
	}

	fmt.Println("")
	result, _ := pterm.InteractiveContinuePrinter{
		DefaultValueIndex: 0,
		DefaultText:       "Unsynced changes found in config. Do you want to sync?",
		TextStyle:         &pterm.ThemeDefault.PrimaryStyle,
		Options:           []string{"y", "n"},
		OptionsStyle:      &pterm.ThemeDefault.SuccessMessageStyle,
		SuffixStyle:       &pterm.ThemeDefault.SecondaryStyle,
		Delimiter:         ": ",
	}.Show()

	if result == "y" {
		for _, manager := range ms {
			tx := a.State.Begin(ctx)
			defer tx.Rollback()

			uw, err := manager.SyncDependencies(ctx, depStatus[manager.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = tx.UpdateDependencyState(ctx, manager.Name(), depsState[manager.Name()])
			if err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return err
			}

			tx = a.State.Begin(ctx)
			defer tx.Rollback()

			uw, err = manager.SyncPackages(ctx, pkgStatus[manager.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = tx.UpdatePackageState(ctx, manager.Name(), pkgsState[manager.Name()])
			if err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return err
			}
		}
	}

	return state.Rotate(config.StateRotations)
}

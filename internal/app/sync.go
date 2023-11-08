package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/core"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/pterm/pterm"
)

func (a App) Sync(ctx context.Context, pms []shared.PackageManager) (err error) {
	depStatus, pkgStatus, err := a.ListStatus(ctx, pms)
	if err != nil {
		return err
	}

	pkgsState := core.UpdatedPackageState(pms, pkgStatus)
	depsState := core.UpdatedDependencyState(pms, depStatus)

	core.PrintPackageList(depStatus, pkgStatus)

	if core.CountUpdatedPkgs(pms, pkgStatus) == 0 && core.CountUpdatedDeps(pms, depStatus) == 0 {
		tx := a.State.Begin(ctx)
		defer tx.Rollback()
		for _, pm := range pms {
			err := a.State.UpdatePackageState(ctx, tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}

			err = a.State.UpdateDependencyState(ctx, tx, pm.Name(), depsState[pm.Name()])
			if err != nil {
				return err
			}
		}
		if err := tx.Commit().Error; err != nil {
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
		for _, pm := range pms {
			tx := a.State.Begin(ctx)
			defer tx.Rollback()

			uw, err := pm.SyncDependencies(ctx, depStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = a.State.UpdateDependencyState(ctx, tx, pm.Name(), depsState[pm.Name()])
			if err != nil {
				return err
			}

			if err := tx.Commit().Error; err != nil {
				return err
			}

			tx = a.State.Begin(ctx)
			defer tx.Rollback()

			uw, err = pm.SyncPackages(ctx, pkgStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = a.State.UpdatePackageState(ctx, tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}

			if err := tx.Commit().Error; err != nil {
				return err
			}
		}
	}

	return state.Rotate(config.StateRotations)
}

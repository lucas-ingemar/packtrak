package machinery

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/pterm/pterm"
)

func Sync(ctx context.Context, pms []shared.PackageManager) (err error) {
	tx := state.Begin()
	defer tx.Rollback()

	depStatus, pkgStatus, err := ListStatus(ctx, tx, pms)
	if err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	pkgsState := UpdatedPackageState(pms, pkgStatus)
	depsState := UpdatedDependencyState(pms, depStatus)

	PrintPackageList(depStatus, pkgStatus)

	if len(TotalUpdatedPkgs(pms, pkgStatus)) == 0 && len(TotalUpdatedDeps(pms, depStatus)) == 0 {
		tx := state.Begin()
		defer tx.Rollback()
		for _, pm := range pms {
			err := state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}

			err = state.UpdateDependencyState(tx, pm.Name(), depsState[pm.Name()])
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
			tx := state.Begin()
			defer tx.Rollback()

			uw, err := pm.SyncDependencies(ctx, depStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = state.UpdateDependencyState(tx, pm.Name(), depsState[pm.Name()])
			if err != nil {
				return err
			}

			if err := tx.Commit().Error; err != nil {
				return err
			}

			tx = state.Begin()
			defer tx.Rollback()

			uw, err = pm.SyncPackages(ctx, pkgStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
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
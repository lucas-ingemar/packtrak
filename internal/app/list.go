package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

func (a *App) ListStatus(ctx context.Context, managerNames []shared.ManagerName) (status.Status, error) {
	ms, err := a.Managers.GetManagers(managerNames)
	if err != nil {
		return status.Status{}, err
	}

	if !a.mustDoSudo(ctx, managerNames, shared.CommandList) {
		return status.Status{}, errors.New("sudo access not granted")
	}

	g, ctx := errgroup.WithContext(ctx)

	multi := pterm.DefaultMultiPrinter
	statusObj := status.Status{}

	for _, manager := range ms {
		spinnerDep, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start(fmt.Sprintf("Listing %s dependencies...", manager.Name()))
		spinnerDep.SuccessPrinter = &shared.PtermInstalled
		spinnerPkg, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start(fmt.Sprintf("Listing %s packages...", manager.Name()))
		spinnerPkg.SuccessPrinter = &shared.PtermInstalled

		func(manager managers.Manager) {
			g.Go(func() error {
				packages, dependencies, err := manifest.Filter(a.Manifest.Pm(manager.Name()))
				if err != nil {
					return err
				}

				packages = lo.Uniq(packages)
				dependencies = lo.Uniq(dependencies)

				stateDeps, err := a.State.GetDependencyState(ctx, manager.Name())
				if err != nil {
					return err
				}

				depStatus, err := manager.ListDependencies(ctx, dependencies, stateDeps)
				if err != nil {
					return err
				}

				statusObj.AddDependencies(manager.Name(), depStatus)
				spinnerDep.Success(fmt.Sprintf("%s dependencies listed", manager.Name()))

				statePkgs, err := a.State.GetPackageState(ctx, manager.Name())
				if err != nil {
					return err
				}

				pkgStatus, err := manager.ListPackages(ctx, packages, statePkgs)
				if err != nil {
					return err
				}

				statusObj.AddPackages(manager.Name(), pkgStatus)
				spinnerPkg.Success(fmt.Sprintf("%s packages listed", manager.Name()))

				time.Sleep(200 * time.Millisecond)
				return nil
			})
		}(manager)
	}
	multi.Start()

	return statusObj, g.Wait()
}

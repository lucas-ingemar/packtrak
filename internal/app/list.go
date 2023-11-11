package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/samber/lo"
)

func (a *App) ListStatus(ctx context.Context, managerNames []shared.ManagerName) (status.Status, error) {
	ms, err := a.Managers.GetManagers(managerNames)
	if err != nil {
		return status.Status{}, err
	}

	if !a.mustDoSudo(ctx, managerNames, shared.CommandList) {
		panic("sudo access not granted")
	}

	statusObj := status.Status{}
	for _, manager := range ms {
		packages, dependencies, err := manifest.Filter(a.Manifest.Pm(manager.Name()))
		if err != nil {
			return status.Status{}, err
		}

		packages = lo.Uniq(packages)
		dependencies = lo.Uniq(dependencies)

		stateDeps, err := a.State.GetDependencyState(ctx, manager.Name())
		if err != nil {
			return status.Status{}, err
		}

		fmt.Printf("Listing %s dependencies...\n", manager.Name())
		depStatus, err := manager.ListDependencies(ctx, dependencies, stateDeps)
		if err != nil {
			return status.Status{}, err
		}
		statusObj.AddDependencies(manager.Name(), depStatus)

		statePkgs, err := a.State.GetPackageState(ctx, manager.Name())
		if err != nil {
			return status.Status{}, err
		}

		fmt.Printf("Listing %s packages...\n", manager.Name())
		pkgStatus, err := manager.ListPackages(ctx, packages, statePkgs)
		if err != nil {
			return status.Status{}, err
		}
		statusObj.AddPackages(manager.Name(), pkgStatus)
	}
	return statusObj, nil
}

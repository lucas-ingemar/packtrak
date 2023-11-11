package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func (a *App) ListStatus(ctx context.Context, managerNames []shared.ManagerName) (map[shared.ManagerName]shared.DependenciesStatus, map[shared.ManagerName]shared.PackageStatus, error) {
	ms, err := a.Managers.GetManagers(managerNames)
	if err != nil {
		return nil, nil, err
	}

	if !a.mustDoSudo(ctx, managerNames, shared.CommandList) {
		panic("sudo access not granted")
	}

	depStatus := map[shared.ManagerName]shared.DependenciesStatus{}
	pkgStatus := map[shared.ManagerName]shared.PackageStatus{}
	for _, manager := range ms {
		packages, dependencies, err := manifest.Filter(a.Manifest.Pm(manager.Name()))
		if err != nil {
			return nil, nil, err
		}

		packages = lo.Uniq(packages)
		dependencies = lo.Uniq(dependencies)

		stateDeps, err := a.State.GetDependencyState(ctx, manager.Name())
		if err != nil {
			return nil, nil, err
		}

		fmt.Printf("Listing %s dependencies...\n", manager.Name())
		depStatus[manager.Name()], err = manager.ListDependencies(ctx, dependencies, stateDeps)
		if err != nil {
			return nil, nil, err
		}

		statePkgs, err := a.State.GetPackageState(ctx, manager.Name())
		if err != nil {
			return nil, nil, err
		}

		fmt.Printf("Listing %s packages...\n", manager.Name())
		pkgStatus[manager.Name()], err = manager.ListPackages(ctx, packages, statePkgs)
		if err != nil {
			return nil, nil, err
		}
	}
	return depStatus, pkgStatus, nil
}

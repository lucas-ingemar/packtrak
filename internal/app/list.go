package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func (a App) ListStatus(ctx context.Context, pms []shared.PackageManager) (map[string]shared.DependenciesStatus, map[string]shared.PackageStatus, error) {
	tx := a.State.Begin(ctx)
	defer tx.Rollback()

	depStatus := map[string]shared.DependenciesStatus{}
	pkgStatus := map[string]shared.PackageStatus{}
	for _, pm := range pms {
		packages, dependencies, err := manifest.Filter(*manifest.Manifest.Pm(pm.Name()))
		if err != nil {
			return nil, nil, err
		}

		packages = lo.Uniq(packages)
		dependencies = lo.Uniq(dependencies)

		stateDeps, err := a.State.GetDependencyState(ctx, tx, pm.Name())
		if err != nil {
			return nil, nil, err
		}

		fmt.Printf("Listing %s dependencies...\n", pm.Name())
		depStatus[pm.Name()], err = pm.ListDependencies(ctx, dependencies, stateDeps)
		if err != nil {
			return nil, nil, err
		}

		statePkgs, err := a.State.GetPackageState(ctx, tx, pm.Name())
		if err != nil {
			return nil, nil, err
		}

		fmt.Printf("Listing %s packages...\n", pm.Name())
		pkgStatus[pm.Name()], err = pm.ListPackages(ctx, packages, statePkgs)
		if err != nil {
			return nil, nil, err
		}
	}
	return depStatus, pkgStatus, nil
}

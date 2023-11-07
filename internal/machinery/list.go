package machinery

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func ListStatus(ctx context.Context, tx *gorm.DB, pms []shared.PackageManager) (map[string]shared.DependenciesStatus, map[string]shared.PackageStatus, error) {
	depStatus := map[string]shared.DependenciesStatus{}
	pkgStatus := map[string]shared.PackageStatus{}
	for _, pm := range pms {
		packages, dependencies, err := manifest.Filter(*manifest.Manifest.Pm(pm.Name()))
		if err != nil {
			return nil, nil, err
		}

		packages = lo.Uniq(packages)
		dependencies = lo.Uniq(dependencies)

		fmt.Printf("Listing %s dependencies...\n", pm.Name())
		depStatus[pm.Name()], err = pm.ListDependencies(ctx, tx, dependencies)
		if err != nil {
			return nil, nil, err
		}
		fmt.Printf("Listing %s packages...\n", pm.Name())
		pkgStatus[pm.Name()], err = pm.ListPackages(ctx, tx, packages)
		if err != nil {
			return nil, nil, err
		}
	}
	return depStatus, pkgStatus, nil
}

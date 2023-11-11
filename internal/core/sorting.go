package core

import (
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func FilterIncomingObjects(pkgs []string, pmManifest manifest.PmManifest, mType manifest.ManifestObjectType) (filteredObjs []string, err error) {
	pkgs = lo.Uniq(pkgs)
	for _, arg := range pkgs {
		var objs []string
		pkgs, deps, err := manifest.Filter(pmManifest)
		if err != nil {
			return nil, err
		}
		if mType == manifest.TypeDependency {
			objs = deps
		} else {
			objs = pkgs
		}
		if lo.Contains(objs, arg) {
			shared.PtermWarning.Printfln("'%s' is already present in manifest", arg)
			continue
		}
		filteredObjs = append(filteredObjs, arg)
	}
	return filteredObjs, nil
}

// func CountUpdatedDeps(pms []managers.Manager, depStatus map[shared.ManagerName]shared.DependenciesStatus) (totUpdatedDeps int) {
// 	for _, pm := range pms {
// 		totUpdatedDeps += len(depStatus[pm.Name()].Missing)
// 		totUpdatedDeps += len(depStatus[pm.Name()].Updated)
// 		totUpdatedDeps += len(depStatus[pm.Name()].Removed)
// 	}
// 	return
// }

// func CountUpdatedPkgs(pms []managers.Manager, pkgStatus map[shared.ManagerName]shared.PackageStatus) (totUpdatedPkgs int) {
// 	for _, pm := range pms {
// 		totUpdatedPkgs += len(pkgStatus[pm.Name()].Missing)
// 		totUpdatedPkgs += len(pkgStatus[pm.Name()].Updated)
// 		totUpdatedPkgs += len(pkgStatus[pm.Name()].Removed)
// 	}
// 	return
// }

// func UpdatedPackageState1(pms []managers.Manager, pkgStatus map[shared.ManagerName]shared.PackageStatus) map[shared.ManagerName][]shared.Package {
// 	pkgsState := map[shared.ManagerName][]shared.Package{}
// 	for _, pm := range pms {
// 		pkgsState[pm.Name()] = []shared.Package{}
// 		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Synced...)
// 		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Updated...)
// 		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Missing...)
// 	}
// 	return pkgsState
// }

// func UpdatedDependencyState1(pms []managers.Manager, depStatus map[shared.ManagerName]shared.DependenciesStatus) map[shared.ManagerName][]shared.Dependency {
// 	depsState := map[shared.ManagerName][]shared.Dependency{}
// 	for _, pm := range pms {
// 		depsState[pm.Name()] = []shared.Dependency{}
// 		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Synced...)
// 		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Updated...)
// 		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Missing...)
// 	}
// 	return depsState
// }

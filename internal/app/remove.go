package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func (a App) Remove(ctx context.Context, apkgs []string, pm shared.PackageManager, mType manifest.ManifestObjectType) error {
	apkgs = lo.Uniq(apkgs)

	pmManifest := a.Manifest.Pm(pm.Name())

	objsToRemove := []string{}
	warningPrinted := false

	var objs []string
	pkgs, deps, err := manifest.Filter(pmManifest)
	if err != nil {
		return err
	}

	if mType == manifest.TypeDependency {
		objs = pm.GetDependencyNames(ctx, deps)
	} else {
		objs = pm.GetPackageNames(ctx, pkgs)
	}

	for _, arg := range apkgs {
		if !lo.Contains(objs, arg) {
			shared.PtermWarning.Printfln("'%s' is not present in manifest", arg)
			warningPrinted = true
			continue
		}
		objsToRemove = append(objsToRemove, arg)
	}

	var toRemove, userWarnings []string

	if mType == manifest.TypeDependency {
		toRemove, userWarnings, err = pm.RemoveDependencies(ctx, deps, objsToRemove)
		if err != nil {
			return err
		}
	} else {
		toRemove, userWarnings, err = pm.RemovePackages(ctx, pkgs, objsToRemove)
		if err != nil {
			return err
		}
	}

	for _, uw := range userWarnings {
		shared.PtermWarning.Println(uw)
		warningPrinted = true
	}

	if warningPrinted {
		fmt.Println("")
	}

	if err = a.Manifest.RemoveGlobal(mType, pm.Name(), toRemove); err != nil {
		return nil
	}
	// if removeDependency {
	// 	pmManifest.Global.RemoveDependencies(toRemove)
	// } else {
	// 	pmManifest.Global.RemovePackages(toRemove)
	// }

	for _, c := range pmManifest.Conditional {
		match, err := manifest.MatchConditional(c)
		if err != nil {
			return err
		}
		if match {
			if err = a.Manifest.RemoveConditional(mType, pm.Name(), c.Type, c.Value, toRemove); err != nil {
				return err
			}
			// if removeDependency {
			// 	pmManifest.Conditional[idx].RemoveDependencies(toRemove)
			// } else {
			// 	pmManifest.Conditional[idx].RemovePackages(toRemove)
			// }
		}
	}

	if err = a.Sync(ctx, []shared.PackageManager{pm}); err != nil {
		return err
	}

	return a.Manifest.Save(config.ManifestFile)
}

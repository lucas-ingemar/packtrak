package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func (a *App) RemoveValidArgsFunc(ctx context.Context, toComplete string, managerName shared.ManagerName, mType manifest.ManifestObjectType) ([]string, error) {
	manager, err := a.Managers.GetManager(managerName)
	if err != nil {
		return nil, err
	}

	pmManifest := a.Manifest.Pm(managerName)
	pkgs, deps, err := manifest.Filter(pmManifest)
	if err != nil {
		return nil, err
	}

	if mType != manifest.TypeDependency {
		return lo.Filter(manager.GetPackageNames(ctx, pkgs),
			func(item string, index int) bool {
				return strings.HasPrefix(item, toComplete)
			}), nil
	} else {
		return lo.Filter(manager.GetDependencyNames(ctx, deps),
			func(item string, index int) bool {
				return strings.HasPrefix(item, toComplete)
			}), nil
	}
}

func (a *App) Remove(ctx context.Context, apkgs []string, managerName shared.ManagerName, mType manifest.ManifestObjectType) error {
	manager, error := a.Managers.GetManager(managerName)
	if error != nil {
		return error
	}

	if !a.mustDoSudo(ctx, []shared.ManagerName{managerName}, shared.CommandRemove) {
		return errors.New("sudo access not granted")
	}

	apkgs = lo.Uniq(apkgs)

	pmManifest := a.Manifest.Pm(manager.Name())

	objsToRemove := []string{}
	warningPrinted := false

	var objs []string
	pkgs, deps, err := manifest.Filter(pmManifest)
	if err != nil {
		return err
	}

	if mType == manifest.TypeDependency {
		objs = manager.GetDependencyNames(ctx, deps)
	} else {
		objs = manager.GetPackageNames(ctx, pkgs)
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
		toRemove, userWarnings, err = manager.RemoveDependencies(ctx, deps, objsToRemove)
		if err != nil {
			return err
		}
	} else {
		toRemove, userWarnings, err = manager.RemovePackages(ctx, pkgs, objsToRemove)
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

	if err = a.Manifest.RemoveGlobal(mType, managerName, toRemove); err != nil {
		return nil
	}

	for _, c := range pmManifest.Conditional {
		match, err := manifest.MatchConditional(c)
		if err != nil {
			return err
		}
		if match {
			if err = a.Manifest.RemoveConditional(mType, managerName, c.Type, c.Value, toRemove); err != nil {
				return err
			}
		}
	}

	if err = a.Sync(ctx, []shared.ManagerName{manager.Name()}); err != nil {
		return err
	}

	return a.Manifest.Save(config.ManifestFile)
}

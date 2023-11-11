package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/core"
	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
)

func (a App) Install(ctx context.Context, apkgs []string, managerName managers.ManagerName, mType manifest.ManifestObjectType, host bool, group string) error {
	manager, error := a.Managers.GetManager(managerName)
	if error != nil {
		return error
	}

	if !shared.MustDoSudo(ctx, []managers.Manager{manager}, shared.CommandInstall) {
		return errors.New("sudo access not granted")
	}

	pmManifest := a.Manifest.Pm(managerName)
	warningPrinted := false

	objsToAdd, err := core.FilterIncomingObjects(apkgs, pmManifest, mType)
	if err != nil {
		return err
	}

	var toAdd, userWarnings []string

	if mType == manifest.TypeDependency {
		toAdd, userWarnings, err = manager.AddDependencies(ctx, objsToAdd)
		if err != nil {
			return err
		}
	} else {
		toAdd, userWarnings, err = manager.AddPackages(ctx, objsToAdd)
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

	//FIXME: This is not very nice, but it works
	if host {
		if err := a.Manifest.AddToHost(toAdd, managerName, mType); err != nil {
			return err
		}
	} else if group != "" {
		if err := a.Manifest.AddToGroup(toAdd, group, managerName, mType); err != nil {
			return err
		}
	} else {
		if err = a.Manifest.AddGlobal(manifest.TypeDependency, managerName, toAdd); err != nil {
			return nil
		}
	}

	if err = a.Sync(ctx, []managers.ManagerName{manager.Name()}); err != nil {
		return err
	}

	return a.Manifest.Save(config.ManifestFile)
}
func (a App) InstallValidArgsFunc(ctx context.Context, managerName managers.ManagerName, toComplete string, mType manifest.ManifestObjectType) (pkgs []string, err error) {
	manager, err := a.Managers.GetManager(managerName)
	if err != nil {
		return nil, err
	}

	return manager.InstallValidArgs(ctx, toComplete, mType == manifest.TypeDependency)
}

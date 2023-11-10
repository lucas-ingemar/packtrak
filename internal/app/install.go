package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/core"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
)

func (a App) Install(ctx context.Context, apkgs []string, pm shared.PackageManager, mType manifest.ManifestObjectType, host bool, group string) error {
	pmManifest := a.Manifest.Pm(pm.Name())
	warningPrinted := false

	objsToAdd, err := core.FilterIncomingObjects(apkgs, pmManifest, mType)
	if err != nil {
		return err
	}

	var toAdd, userWarnings []string

	if mType == manifest.TypeDependency {
		toAdd, userWarnings, err = pm.AddDependencies(ctx, objsToAdd)
		if err != nil {
			return err
		}
	} else {
		toAdd, userWarnings, err = pm.AddPackages(ctx, objsToAdd)
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
		if err := a.Manifest.AddToHost(toAdd, pm.Name(), mType); err != nil {
			return err
		}
	} else if group != "" {
		if err := a.Manifest.AddToGroup(toAdd, group, pm.Name(), mType); err != nil {
			return err
		}
	} else {
		if err = a.Manifest.AddGlobal(manifest.TypeDependency, pm.Name(), toAdd); err != nil {
			return nil
		}
		// if installDependency {
		// 	a.Manifest.Pm(pm.Name()).Global.AddDependencies(toAdd)
		// } else {
		// 	a.Manifest.Pm(pm.Name()).Global.AddPackages(toAdd)
		// }
	}

	if err = a.Sync(ctx, []shared.PackageManager{pm}); err != nil {
		return err
	}

	return a.Manifest.Save(config.ManifestFile)
}

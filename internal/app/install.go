package app

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/core"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
)

func (a App) Install(ctx context.Context, apkgs []string, pm shared.PackageManager, pmManifest *shared.PmManifest, installDependency bool, host bool, group string) error {
	warningPrinted := false

	objsToAdd, err := core.FilterIncomingObjects(apkgs, *pmManifest, installDependency)
	if err != nil {
		return err
	}

	var toAdd, userWarnings []string

	if installDependency {
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
		if err := manifest.AddToHost(toAdd, pm.Name(), installDependency); err != nil {
			return err
		}
	} else if group != "" {
		if err := manifest.AddToGroup(toAdd, pm.Name(), group, installDependency); err != nil {
			return err
		}
	} else {
		if installDependency {
			manifest.Manifest.Pm(pm.Name()).Global.AddDependencies(toAdd)
		} else {
			manifest.Manifest.Pm(pm.Name()).Global.AddPackages(toAdd)
		}
	}

	if err = a.Sync(ctx, []shared.PackageManager{pm}); err != nil {
		return err
	}

	return manifest.SaveManifest(config.ManifestFile)
}

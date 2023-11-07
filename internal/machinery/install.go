package machinery

import (
	"context"
	"fmt"
	"os"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func Install(ctx context.Context, apkgs []string, pm shared.PackageManager, pmManifest *shared.PmManifest, installDependency bool, host bool, group string) error {
	objsToAdd := []string{}
	warningPrinted := false

	apkgs = lo.Uniq(apkgs)

	for _, arg := range apkgs {
		var objs []string
		pkgs, deps, err := manifest.Filter(*pmManifest)
		if err != nil {
			return err
		}
		if installDependency {
			objs = deps
		} else {
			objs = pkgs
		}
		if lo.Contains(objs, arg) {
			shared.PtermWarning.Printfln("'%s' is already present in manifest", arg)
			warningPrinted = true
			continue
		}
		objsToAdd = append(objsToAdd, arg)
	}

	var toAdd, userWarnings []string
	var err error

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
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		mc, err := manifest.Manifest.Pm(pm.Name()).GetOrAddConditional(shared.MConditionHost, hostname)
		if err != nil {
			return err
		}
		if installDependency {
			mc.AddDependencies(toAdd)
		} else {
			mc.AddPackages(toAdd)
		}
	} else if group != "" {
		mc, err := manifest.Manifest.Pm(pm.Name()).GetOrAddConditional(shared.MConditionGroup, group)
		if err != nil {
			return err
		}
		if installDependency {
			mc.AddDependencies(toAdd)
		} else {
			mc.AddPackages(toAdd)
		}
	} else {
		if installDependency {
			manifest.Manifest.Pm(pm.Name()).Global.AddDependencies(toAdd)
		} else {
			manifest.Manifest.Pm(pm.Name()).Global.AddPackages(toAdd)
		}
	}

	err = Sync(ctx, []shared.PackageManager{pm})
	if err != nil {
		return err
	}

	return manifest.SaveManifest(config.ManifestFile)
}

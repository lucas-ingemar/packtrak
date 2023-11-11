package app

import (
	"context"

	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
)

type AppFace interface {
	Install(ctx context.Context, apkgs []string, managerName managers.ManagerName, mType manifest.ManifestObjectType, host bool, group string) error
	InstallValidArgsFunc(ctx context.Context, managerName managers.ManagerName, toComplete string, mType manifest.ManifestObjectType) (pkgs []string, err error)
	ListStatus(ctx context.Context, managerNames []managers.ManagerName) (map[managers.ManagerName]shared.DependenciesStatus, map[managers.ManagerName]shared.PackageStatus, error)
	Remove(ctx context.Context, apkgs []string, managerName managers.ManagerName, mType manifest.ManifestObjectType) error
	RemoveValidArgsFunc(ctx context.Context, toComplete string, managerName managers.ManagerName, mType manifest.ManifestObjectType) ([]string, error)
	Sync(ctx context.Context, managerNames []managers.ManagerName) (err error)
	PrintPackageList(depStatus map[managers.ManagerName]shared.DependenciesStatus, pkgStatus map[managers.ManagerName]shared.PackageStatus) error
	//FIXME: Might want to create a typ for manager names
	ListManagers() []managers.ManagerName
	// GetManifest() manifest.ManifestFace
}

type App struct {
	Managers managers.ManagerFactoryFace
	Manifest manifest.ManifestFace
	State    state.StateFace
}

func (a App) ListManagers() []managers.ManagerName {
	return a.Managers.ListManagers()
}

// func (a *App) GetManifest() manifest.ManifestFace {
// 	return a.Manifest
// }

func NewApp(managers managers.ManagerFactoryFace, manifest manifest.ManifestFace, state state.StateFace) App {
	return App{
		Managers: managers,
		Manifest: manifest,
		State:    state,
	}
}

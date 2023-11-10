package app

import (
	"context"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
)

type AppFace interface {
	Install(ctx context.Context, apkgs []string, pm shared.PackageManager, mType manifest.ManifestObjectType, host bool, group string) error
	ListStatus(ctx context.Context, pms []shared.PackageManager) (map[string]shared.DependenciesStatus, map[string]shared.PackageStatus, error)
	Remove(ctx context.Context, apkgs []string, pm shared.PackageManager, mType manifest.ManifestObjectType) error
	Sync(ctx context.Context, pms []shared.PackageManager) (err error)
	// GetManifest() manifest.ManifestFace
}

type App struct {
	Manifest manifest.ManifestFace
	State    state.StateFace
}

// func (a *App) GetManifest() manifest.ManifestFace {
// 	return a.Manifest
// }

func NewApp(manifest manifest.ManifestFace, state state.StateFace) App {
	return App{
		Manifest: manifest,
		State:    state,
	}
}

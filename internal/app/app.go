package app

import (
	"context"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
)

type AppFace interface {
	Install(ctx context.Context, apkgs []string, pm shared.PackageManager, pmManifest *shared.PmManifest, installDependency bool, host bool, group string) error
	ListStatus(ctx context.Context, pms []shared.PackageManager) (map[string]shared.DependenciesStatus, map[string]shared.PackageStatus, error)
	Remove(ctx context.Context, apkgs []string, pm shared.PackageManager, pmManifest *shared.PmManifest, removeDependency bool) error
	Sync(ctx context.Context, pms []shared.PackageManager) (err error)
}

type App struct {
	State state.StateFace
}

func NewApp(state state.StateFace) App {
	return App{
		State: state,
	}
}

package shared

import (
	"context"

	"gorm.io/gorm"
)

type PackageManager interface {
	Name() string
	Icon() string
	NeedsSudo() []CommandName

	GetPackageNames(ctx context.Context, packages []string) []string
	GetDependencyNames(ctx context.Context, deps []string) []string

	// FIXME: Update to new format ???? This looks fucked up
	AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error)
	AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error)
	InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error)
	ListDependencies(ctx context.Context, tx *gorm.DB, deps []string) (depStatus DependenciesStatus, err error)
	ListPackages(ctx context.Context, tx *gorm.DB, packages []string) (packageStatus PackageStatus, err error)
	// FIXME: Update to new format ???? This looks fucked up
	RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error)
	RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error)
	SyncDependencies(ctx context.Context, depStatus DependenciesStatus) (userWarnings []string, err error)
	SyncPackages(ctx context.Context, packageStatus PackageStatus) (userWarnings []string, err error)
}

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

	// FIXME: Update to new format ???? This looks fucked up
	Add(ctx context.Context, packages []string, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error)
	InstallValidArgs(ctx context.Context, toComplete string) ([]string, error)
	ListDependencies(ctx context.Context, tx *gorm.DB, deps []string) (depStatus DependenciesStatus, err error)
	ListPackages(ctx context.Context, tx *gorm.DB, packages []string) (packageStatus PackageStatus, err error)
	// FIXME: Update to new format ???? This looks fucked up
	Remove(ctx context.Context, packages []string, pkgs []string) (packagesToRemove []string, userWarnings []string, err error)
	SyncDependencies(ctx context.Context, depStatus DependenciesStatus) (userWarnings []string, err error)
	SyncPackages(ctx context.Context, packageStatus PackageStatus) (userWarnings []string, err error)
}

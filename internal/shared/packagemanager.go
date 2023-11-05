package shared

import (
	"context"

	"gorm.io/gorm"
)

type PackageManager interface {
	Name() string
	Icon() string
	NeedsSudo() []CommandName

	GetPackageNames(ctx context.Context, packagesConfig PmPackages) []string

	// FIXME: Update to new format
	Add(ctx context.Context, packagesConfig PmPackages, pkgs []string) (packagesConfigUpdated PmPackages, userWarnings []string, err error)
	InstallValidArgs(ctx context.Context, toComplete string) ([]string, error)
	ListDependencies(ctx context.Context, tx *gorm.DB, packages PmPackages) (depStatus DependenciesStatus, err error)
	ListPackages(ctx context.Context, tx *gorm.DB, packages PmPackages) (packageStatus PackageStatus, err error)
	// FIXME: Update to new format
	Remove(ctx context.Context, packagesConfig PmPackages, pkgs []string) (packagesConfigUpdated PmPackages, userWarnings []string, err error)
	SyncDependencies(ctx context.Context, depStatus DependenciesStatus) (userWarnings []string, err error)
	SyncPackages(ctx context.Context, packageStatus PackageStatus) (userWarnings []string, err error)
}

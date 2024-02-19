package github

import (
	"context"
	"fmt"
	"os"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/spf13/viper"
)

const Name shared.ManagerName = "github"

const (
	packageDirectoryKey = "package_directory"
)

func New() *Github {
	return &Github{}
}

type Github struct {
	pkgDirectory string
}

func (gh *Github) Name() shared.ManagerName {
	return Name
}

func (gh *Github) Icon() string {
	return ""
}

func (gh *Github) ShortDesc() string {
	return "Manage Github released files"
}

func (gh *Github) LongDesc() string {
	return "Manage Github released files. Download and keep track of artifacts from releases from github"
}

func (gh *Github) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{}
}

func (gh *Github) InitConfig() {
	viper.SetDefault(shared.ConfigKeyName(Name, packageDirectoryKey), "")
}

func (gh *Github) InitCheckCmd() error {
	return nil
}

func (gh *Github) InitCheckConfig() error {
	gh.pkgDirectory = viper.GetString(shared.ConfigKeyName(Name, packageDirectoryKey))
	if gh.pkgDirectory == "" {
		return fmt.Errorf("config '%s' must be set", packageDirectoryKey)
	}

	fInfo, err := os.Stat(gh.pkgDirectory)
	if err != nil {
		return err
	}

	if !fInfo.IsDir() {
		return fmt.Errorf("'%s' is not pointing to a directory", packageDirectoryKey)
	}

	return nil
}

func (gh *Github) GetPackageNames(ctx context.Context, packages []string) []string {
	// FIXME
	return nil
}

func (gh *Github) GetDependencyNames(ctx context.Context, deps []string) []string {
	return nil
}

func (gh *Github) InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error) {
	// FIXME
	return nil, nil
}

func (gh *Github) AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error) {
	// FIXME
	return
}

func (gh *Github) AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (gh *Github) ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus status.DependenciesStatus, err error) {
	return
}

func (gh *Github) ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus status.PackageStatus, err error) {
	// FIXME
	return
}

func (gh *Github) RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error) {
	// FIXME
	return
}

func (gh *Github) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (gh *Github) SyncDependencies(ctx context.Context, depStatus status.DependenciesStatus) (userWarnings []string, err error) {
	return
}

func (gh *Github) SyncPackages(ctx context.Context, packageStatus status.PackageStatus) (userWarnings []string, err error) {
	// FIXME
	return
}

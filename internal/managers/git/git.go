package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

func New() *Git {
	return &Git{
		CommandExecutorFace: commandExecutor{},
	}
}

const Name shared.ManagerName = "git"

const (
	packageDirectoryKey        = "package_directory"
	includeUnstableReleasesKey = "include_unstable_releases"
)

type Git struct {
	pkgDirectory            string
	includeUnstableReleases bool
	CommandExecutorFace
}

func (g *Git) Name() shared.ManagerName {
	return Name
}

func (g *Git) Icon() string {
	return "󰊢"
}

func (g *Git) ShortDesc() string {
	return "Clone git repositories"
}

func (g *Git) LongDesc() string {
	return ""
}

func (g *Git) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{}
}

func (g *Git) InitCheckCmd() error {
	return nil
}

func (g *Git) InitCheckConfig() error {
	g.pkgDirectory = viper.GetString(shared.ConfigKeyName(Name, packageDirectoryKey))
	if g.pkgDirectory == "" {
		return fmt.Errorf("config '%s' must be set", packageDirectoryKey)
	}

	fInfo, err := os.Stat(g.pkgDirectory)
	if err != nil {
		return err
	}

	if !fInfo.IsDir() {
		return fmt.Errorf("'%s' is not pointing to a directory", packageDirectoryKey)
	}

	g.includeUnstableReleases = viper.GetBool(shared.ConfigKeyName(Name, includeUnstableReleasesKey))

	return nil
}

func (g *Git) InitConfig() {
	viper.SetDefault(shared.ConfigKeyName(Name, packageDirectoryKey), "")
}

func (g *Git) GetPackageNames(ctx context.Context, packages []string) []string {
	return nil
}

func (g *Git) GetDependencyNames(ctx context.Context, deps []string) []string {
	return nil
}

func (g *Git) InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error) {
	return nil, nil
}

func (g *Git) AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error) {
	return pkgsToAdd, nil, nil
}

func (g *Git) AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (g *Git) ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus status.DependenciesStatus, err error) {
	return
}

func (g *Git) ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus status.PackageStatus, err error) {
	installedPkgs, err := g.ListInstalledPkgs(ctx, g.pkgDirectory, g.includeUnstableReleases)
	if err != nil {
		return status.PackageStatus{}, err
	}

	pkgObjs := []shared.Package{}
	for _, pkgName := range packages {
		pkg, err := g.GetRemotePkgMeta(ctx, pkgName, g.includeUnstableReleases)
		if err != nil {
			return status.PackageStatus{}, err
		}
		pkgObjs = append(pkgObjs, pkg)
	}

	for _, pkg := range pkgObjs {
		matchedPkgs := lo.Filter(installedPkgs, func(item shared.Package, _ int) bool {
			return item.FullName == pkg.FullName
		})
		if len(matchedPkgs) > 0 {
			pkg.Version = matchedPkgs[0].Version
			if pkg.Version != pkg.LatestVersion {
				packageStatus.Updated = append(packageStatus.Updated, pkg)
				continue
			}
			packageStatus.Synced = append(packageStatus.Synced, pkg)
		} else {
			packageStatus.Missing = append(packageStatus.Missing, pkg)
		}
	}

	for _, pkg := range statePkgs {
		if !lo.Contains(packages, pkg) {
			pParts := strings.Split(pkg, "/")
			packageStatus.Removed = append(packageStatus.Removed, shared.Package{
				Name:     strings.TrimSuffix(pParts[len(pParts)-1], ".git"),
				FullName: pkg,
			})
		}
	}

	return
}

func (g *Git) RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error) {
	return
}

func (g *Git) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (g *Git) SyncDependencies(ctx context.Context, depStatus status.DependenciesStatus) (userWarnings []string, err error) {
	return
}

func (g *Git) SyncPackages(ctx context.Context, packageStatus status.PackageStatus) (userWarnings []string, err error) {
	for _, pkg := range packageStatus.Missing {
		err = shared.PtermSpinner(shared.PtermSpinnerInstall, pkg.Name, func() error {
			return g.InstallPkg(ctx, pkg, g.pkgDirectory)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}

	for _, pkg := range packageStatus.Updated {
		err = shared.PtermSpinner(shared.PtermSpinnerUpdate, pkg.Name, func() error {
			return g.UpdatePkg(ctx, pkg, g.pkgDirectory)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}

	for _, pkg := range packageStatus.Removed {
		err = shared.PtermSpinner(shared.PtermSpinnerRemove, pkg.Name, func() error {
			return g.RemovePkg(ctx, pkg, g.pkgDirectory)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}

	return
}

package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/rs/zerolog/log"
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
	return "Keep track of your wanted git repos. Will keep repos up to date. If tags are found the latest tag will be checked out, if not the latest commit will be used."
}

func (g *Git) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{}
}

func (g *Git) InitCheckCmd() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return errors.New("'git' command not found on the computer")
	}
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
	viper.SetDefault(shared.ConfigKeyName(Name, includeUnstableReleasesKey), false)
}

func (g *Git) GetPackageNames(ctx context.Context, packages []string) []string {
	var retpkgs []string
	for _, p := range packages {
		retpkgs = append(retpkgs, g.PkgNameFromUrl(p))
	}
	return retpkgs
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
	installedPkgs, err := g.ListInstalledPkgs(ctx, packages, g.pkgDirectory, g.includeUnstableReleases)
	if err != nil {
		return status.PackageStatus{}, err
	}

	pkgObjs := []shared.Package{}
	for _, pkgNameWithTag := range packages {
		pNT := strings.Split(pkgNameWithTag, ":")
		pkgName := pkgNameWithTag
		useHeadRelease := false
		if pNT[len(pNT)-1] == "latest" {
			useHeadRelease = true
			pkgName = strings.Join(pNT[:len(pNT)-1], ":")
		}
		pkg, err := g.GetRemotePkgMeta(ctx, pkgName, g.includeUnstableReleases, useHeadRelease)
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

	pkgsNoTags := lo.Map(packages, func(p string, _ int) string {
		return strings.TrimSuffix(p, ":latest")
	})

	for _, pkg := range statePkgs {
		if !lo.Contains(pkgsNoTags, pkg) {
			packageStatus.Removed = append(packageStatus.Removed, shared.Package{
				Name:     g.PkgNameFromUrl(pkg),
				FullName: pkg,
			})
		}
	}

	return
}

func (g *Git) RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error) {
	for _, p := range pkgsToRemove {
		pkg, err := g.GetBasicPkgInfo(ctx, p, g.pkgDirectory)
		if err != nil {
			return nil, nil, err
		}
		packagesToRemove = append(packagesToRemove, pkg.FullName)
	}
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
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}

	for _, pkg := range packageStatus.Updated {
		err = shared.PtermSpinner(shared.PtermSpinnerUpdate, pkg.Name, func() error {
			return g.UpdatePkg(ctx, pkg, g.pkgDirectory)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}

	for _, pkg := range packageStatus.Removed {
		err = shared.PtermSpinner(shared.PtermSpinnerRemove, pkg.Name, func() error {
			return g.RemovePkg(ctx, pkg, g.pkgDirectory)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}

	return
}

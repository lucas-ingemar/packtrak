package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

const Name shared.ManagerName = "github"

const (
	packageDirectoryKey = "package_directory"
	binDirectoryKey     = "bin_directory"
	symlinkToBinKey     = "symlink_to_bin"
)

func New() *Github {
	return &Github{
		CommandExecutorFace: commandExecutor{
			GithubHttpFace: GithubHttp{},
		},
	}
}

type Github struct {
	CommandExecutorFace

	pkgDirectory string
	binDirectory string
	symlinkToBin bool
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
	viper.SetDefault(shared.ConfigKeyName(Name, binDirectoryKey), "")
	viper.SetDefault(shared.ConfigKeyName(Name, symlinkToBinKey), false)
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

	gh.symlinkToBin = viper.GetBool(shared.ConfigKeyName(Name, symlinkToBinKey))
	gh.binDirectory = viper.GetString(shared.ConfigKeyName(Name, binDirectoryKey))
	if gh.symlinkToBin && gh.binDirectory == "" {
		return fmt.Errorf("%s must be set if %s is true", binDirectoryKey, symlinkToBinKey)
	}

	return nil
}

func (gh *Github) GetPackageNames(ctx context.Context, packages []string) []string {
	var retpkgs []string
	for _, p := range packages {
		user, repo, _, err := url2pkgComponents(p)
		if err != nil {
			retpkgs = append(retpkgs, p)
		} else {
			retpkgs = append(retpkgs, fmt.Sprintf("%s/%s", user, repo))
		}
	}
	return retpkgs
}

func (gh *Github) InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error) {
	return nil, nil
}

func (gh *Github) AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error) {
	lo.ForEach(pkgsToAdd, func(pkgName string, _ int) {
		sanitizedPkgName, err := sanitizeGithubUrl(pkgName)
		if err != nil {
			userWarnings = append(userWarnings, err.Error())
			return
		}
		packagesUpdated = append(packagesUpdated, sanitizedPkgName)
	})
	return
}

func (gh *Github) ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus status.PackageStatus, err error) {
	installedPkgs, err := gh.ListInstalledPkgs(ctx, gh.pkgDirectory)
	if err != nil {
		return status.PackageStatus{}, err
	}

	manifestPkgs, err := gh.GetManifestPackages(ctx, packages)
	if err != nil {
		return status.PackageStatus{}, err
	}

	for _, pkg := range manifestPkgs {
		matchedPkgs := lo.Filter(installedPkgs, func(item shared.Package, _ int) bool {
			return item.Name == pkg.Name
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
			user, repo, _, err := url2pkgComponents(pkg)
			if err != nil {
				return status.PackageStatus{}, err
			}
			packageStatus.Removed = append(packageStatus.Removed, shared.Package{
				Name:     fmt.Sprintf("%s/%s", user, repo),
				FullName: pkg,
			})
		}
	}

	return
}

func (gh *Github) RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error) {
	for _, pR := range pkgsToRemove {
		for _, pA := range allPkgs {
			if strings.Contains(pA, pR) {
				packagesToRemove = append(packagesToRemove, pA)
			}
		}
	}
	return
}

func (gh *Github) SyncPackages(ctx context.Context, packageStatus status.PackageStatus) (userWarnings []string, err error) {
	binPath := ""
	if gh.symlinkToBin {
		binPath = gh.binDirectory
	}

	for _, pkg := range packageStatus.Missing {
		err = shared.PtermSpinner(shared.PtermSpinnerInstall, pkg.Name, func() error {
			return gh.InstallPkg(ctx, pkg, gh.pkgDirectory, binPath)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}

	for _, pkg := range packageStatus.Updated {
		err = shared.PtermSpinner(shared.PtermSpinnerUpdate, pkg.Name, func() error {
			err := gh.RemovePkg(ctx, pkg, gh.pkgDirectory, binPath)
			if err != nil {
				return err
			}
			return gh.InstallPkg(ctx, pkg, gh.pkgDirectory, binPath)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}

	for _, pkg := range packageStatus.Removed {
		err = shared.PtermSpinner(shared.PtermSpinnerRemove, pkg.Name, func() error {
			return gh.RemovePkg(ctx, pkg, gh.pkgDirectory, binPath)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}
	return
}

func (gh *Github) GetDependencyNames(ctx context.Context, deps []string) []string {
	return nil
}

func (gh *Github) AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (gh *Github) ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus status.DependenciesStatus, err error) {
	return
}

func (gh *Github) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (gh *Github) SyncDependencies(ctx context.Context, depStatus status.DependenciesStatus) (userWarnings []string, err error) {
	return
}

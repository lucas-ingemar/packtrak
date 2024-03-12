package flatpak

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

const Name shared.ManagerName = "flatpak"

const (
	userSpaceInstallationKey = "user_space_installations"
)

func New() *Flatpak {
	return &Flatpak{
		CommandExecutorFace: commandExecutor{},
	}
}

type Flatpak struct {
	CommandExecutorFace

	userSpaceInstallation bool
}

func (f *Flatpak) Name() shared.ManagerName {
	return Name
}

func (f *Flatpak) Icon() string {
	return "󰏖"
}

func (f *Flatpak) ShortDesc() string {
	return "Manage Flatpak packages"
}

func (f *Flatpak) LongDesc() string {
	return "Manage Flatpak packages. No need to run flatpak update anymore"
}

func (f *Flatpak) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{}
}

func (f *Flatpak) InitConfig() {
	viper.SetDefault(shared.ConfigKeyName(Name, userSpaceInstallationKey), false)
}

func (f *Flatpak) InitCheckCmd() error {
	_, err := exec.LookPath("flatpak")
	if err != nil {
		return errors.New("'flatpak' command not found on the computer")
	}
	return nil
}

func (f *Flatpak) InitCheckConfig() error {
	f.userSpaceInstallation = viper.GetBool(shared.ConfigKeyName(Name, userSpaceInstallationKey))
	return nil
}

func (f *Flatpak) GetPackageNames(ctx context.Context, packages []string) []string {
	return lo.Map(packages, func(p string, _ int) string {
		return strings.Split(p, ":")[1]
	})
}

func (f *Flatpak) InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error) {
	return nil, nil
}

func (f *Flatpak) AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error) {
	lo.ForEach(pkgsToAdd, func(pkgName string, _ int) {
		err := checkNameFormat(pkgName)
		if err != nil {
			userWarnings = append(userWarnings, err.Error())
			return
		}
		packagesUpdated = append(packagesUpdated, pkgName)
	})
	return
}

func (f *Flatpak) ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus status.PackageStatus, err error) {
	installedPkgs, err := f.ListInstalledPkgs(ctx, f.userSpaceInstallation)
	if err != nil {
		return
	}

	updateablePkgs, err := f.ListUpdateablePkgs(ctx, f.userSpaceInstallation)
	if err != nil {
		return
	}

	for _, pkgFullName := range packages {
		matchedPkgs := lo.Filter(installedPkgs, func(item shared.Package, _ int) bool {
			return item.FullName == pkgFullName
		})

		updateablePkg := lo.Filter(updateablePkgs, func(item shared.Package, _ int) bool {
			return item.FullName == pkgFullName
		})

		if len(matchedPkgs) > 0 {
			pkg := matchedPkgs[0]
			if len(updateablePkg) > 0 {
				packageStatus.Updated = append(packageStatus.Updated, pkg)
				continue
			}
			packageStatus.Synced = append(packageStatus.Synced, pkg)
		} else {
			packageStatus.Missing = append(packageStatus.Missing, shared.Package{
				Name:          pkgFullName,
				FullName:      pkgFullName,
				Version:       "",
				LatestVersion: "",
				RepoUrl:       "",
			})
		}
	}

	for _, pkg := range statePkgs {
		if !lo.Contains(packages, pkg) {
			removedPkg := pkg
			matchedPkgs := lo.Filter(installedPkgs, func(item shared.Package, _ int) bool {
				return item.FullName == pkg
			})

			if len(matchedPkgs) > 0 {
				removedPkg = matchedPkgs[0].Name
			}

			packageStatus.Removed = append(packageStatus.Removed, shared.Package{
				Name:     removedPkg,
				FullName: pkg,
			})
		}
	}
	return
}

func (f *Flatpak) RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error) {
	for _, pR := range pkgsToRemove {
		for _, pA := range allPkgs {
			if strings.Contains(pA, pR) {
				packagesToRemove = append(packagesToRemove, pA)
			}
		}
	}
	return
}

func (f *Flatpak) SyncPackages(ctx context.Context, packageStatus status.PackageStatus) (userWarnings []string, err error) {
	for _, pkg := range packageStatus.Missing {
		err = shared.PtermSpinner(shared.PtermSpinnerInstall, pkg.Name, func() error {
			return f.InstallPkg(ctx, pkg, f.userSpaceInstallation)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}

	for _, pkg := range packageStatus.Updated {
		err = shared.PtermSpinner(shared.PtermSpinnerUpdate, pkg.Name, func() error {
			return f.UpdatePkg(ctx, pkg, f.userSpaceInstallation)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}

	for _, pkg := range packageStatus.Removed {
		err = shared.PtermSpinner(shared.PtermSpinnerRemove, pkg.Name, func() error {
			return f.RemovePkg(ctx, pkg, f.userSpaceInstallation)
		})
		if err != nil {
			log.Err(err).Str("manager", string(Name)).Str("package", pkg.Name)
			err = nil
		}
	}
	return
}

func (f *Flatpak) GetDependencyNames(ctx context.Context, deps []string) []string {
	return nil
}

func (f *Flatpak) AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (f *Flatpak) ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus status.DependenciesStatus, err error) {
	return
}

func (f *Flatpak) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (f *Flatpak) SyncDependencies(ctx context.Context, depStatus status.DependenciesStatus) (userWarnings []string, err error) {
	return
}

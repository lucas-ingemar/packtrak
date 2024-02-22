package dnf

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func New() *Dnf {
	return &Dnf{
		&commandExecutor{},
	}
}

const Name shared.ManagerName = "dnf"

type Dnf struct {
	CommandExecutorFace
}

func (d *Dnf) Name() shared.ManagerName {
	return Name
}

func (d *Dnf) Icon() string {
	return ""
}

func (d *Dnf) ShortDesc() string {
	return "A package manager for RPM-based Linux distributions"
}

func (d *Dnf) LongDesc() string {
	return "DNF is the next upcoming major version of YUM, a package manager for RPM-based Linux distributions. It roughly maintains CLI compatibility with YUM and defines a strict API for extensions and plugins."
}

func (d *Dnf) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{shared.CommandInstall, shared.CommandRemove, shared.CommandSync}
}

func (d *Dnf) InitCheckCmd() error {
	_, err := exec.LookPath("dnf")
	if err != nil {
		return errors.New("'dnf' command not found on the computer")
	}
	return nil
}

func (d *Dnf) InitConfig() {
}

func (d *Dnf) InitCheckConfig() error {
	return nil
}

func (d *Dnf) GetPackageNames(ctx context.Context, packages []string) []string {
	return packages
}

func (d *Dnf) GetDependencyNames(ctx context.Context, deps []string) []string {
	fixedDeps := []string{}
	for _, d := range deps {
		fixedDeps = append(fixedDeps, strings.SplitN(d, ":", 2)[1])
	}
	return fixedDeps
}

func (d *Dnf) AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error) {
	for _, pkg := range pkgsToAdd {
		isSysPkg, err := d.isSystemPackage(ctx, pkg)
		if err != nil {
			return packagesUpdated, []string{}, err
		}

		if !isSysPkg {
			packagesUpdated = append(packagesUpdated, pkg)
		} else {
			userWarnings = append(userWarnings, fmt.Sprintf("'%s' is a system package and cannot be managed", pkg))
		}
	}
	return packagesUpdated, userWarnings, nil
}

func (d *Dnf) AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error) {
	for _, dep := range depsToAdd {
		if strings.HasPrefix(dep, "copr:") || strings.HasPrefix(dep, "cm:") {
			depsUpdated = append(depsUpdated, dep)
		} else {
			userWarnings = append(userWarnings, fmt.Sprintf("Dependency '%s' has an incorrect format", dep))
		}
	}
	return
}

func (d *Dnf) InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error) {
	if dependencies {
		return []string{}, nil
	}

	res, err := shared.Command(ctx, "dnf", []string{"list", "--available", toComplete + "*"}, false, nil)
	if err != nil {
		return nil, err
	}

	dnfList := strings.Split(strings.TrimSpace(res), "\n")
	for idx, line := range dnfList {
		if strings.Contains(line, "Available Packages") {
			dnfList = dnfList[idx+1:]
			break
		}
	}

	pkgs := []string{}

	for _, pkg := range dnfList {
		pkgs = append(pkgs, strings.Split(pkg, ".")[0])
	}

	return pkgs, nil
}

func (d *Dnf) ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus status.DependenciesStatus, err error) {
	installedCoprs, err := d.ListCoprs(ctx)
	if err != nil {
		return status.DependenciesStatus{}, err
	}

	installedCms, err := d.ListCm(ctx)
	if err != nil {
		return status.DependenciesStatus{}, err
	}

	pCoprs, pCms := d.sortDeps(deps)

	// COPR
	for _, dep := range pCoprs {
		depFound := false
		for _, dnfDep := range installedCoprs {
			if dnfDep == d.coprFilename(dep.Name) {
				depStatus.Synced = append(depStatus.Synced, dep)
				depFound = true
				break
			}
		}
		if !depFound {
			depStatus.Missing = append(depStatus.Missing, dep)
		}
	}
	// CM
	for _, dep := range pCms {
		depFound := false
		for _, dnfDep := range installedCms {
			if strings.HasSuffix(dep.Name, dnfDep) {
				depStatus.Synced = append(depStatus.Synced, dep)
				depFound = true
				break
			}
		}
		if !depFound {
			depStatus.Missing = append(depStatus.Missing, dep)
		}
	}

	sCoprs, sCms := d.sortDeps(stateDeps)

	// COPR
	for _, dep := range sCoprs {
		for _, dnfDep := range installedCoprs {
			if dnfDep == d.coprFilename(dep.Name) {
				if !lo.Contains(deps, dep.FullName) {
					depStatus.Removed = append(depStatus.Removed, dep)
				}
				break
			}
		}
	}
	// CM
	for _, dep := range sCms {
		for _, dnfDep := range installedCms {
			if strings.HasSuffix(dep.Name, dnfDep) {
				if !lo.Contains(deps, dep.FullName) {
					depStatus.Removed = append(depStatus.Removed, dep)
				}
				break
			}
		}
	}

	return
}

func (d *Dnf) ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus status.PackageStatus, err error) {
	dnfList, dnfVersions, err := d.ListInstalledPkgs(ctx)
	if err != nil {
		return
	}

	for _, pkg := range packages {
		pkgFound := false
		for idx, dnfPkg := range dnfList {
			if dnfPkg == pkg {
				packageStatus.Synced = append(packageStatus.Synced, shared.Package{Name: pkg, FullName: pkg, Version: dnfVersions[idx]})
				pkgFound = true
				break
			}
		}
		if !pkgFound {
			packageStatus.Missing = append(packageStatus.Missing, shared.Package{Name: pkg, FullName: pkg})
		}
	}

	for _, pkg := range statePkgs {
		for _, dnfPkg := range dnfList {
			if dnfPkg == pkg {
				if !lo.Contains(packages, pkg) {
					packageStatus.Removed = append(packageStatus.Removed, shared.Package{Name: pkg, FullName: pkg})
				}
				break
			}
		}
	}

	return
}

func (d *Dnf) RemovePackages(ctx context.Context, allPkgs []string, pkgs []string) (packagesToRemove []string, userWarnings []string, err error) {
	for _, pkg := range pkgs {
		var isSysPkg bool
		isSysPkg, err = d.isSystemPackage(ctx, pkg)
		if err != nil {
			return
		}

		if isSysPkg {
			userWarnings = append(userWarnings, fmt.Sprintf("'%s' is a system package and cannot be managed", pkg))
			continue
		}
		packagesToRemove = append(packagesToRemove, pkg)
	}

	return
}

func (d *Dnf) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	for _, rDep := range depsToRemove {
		for _, aDep := range allDeps {
			if strings.SplitN(aDep, ":", 2)[1] == rDep {
				depsUpdated = append(depsUpdated, aDep)
				continue
			}
		}
	}
	return
}

func (d *Dnf) SyncDependencies(ctx context.Context, depStatus status.DependenciesStatus) (userWarnings []string, err error) {
	if len(depStatus.Missing) > 0 {
		fmt.Println("")
		mCoprs := []string{}
		mCms := []string{}
		for _, dep := range depStatus.Missing {
			if strings.HasPrefix(dep.FullName, "copr:") {
				mCoprs = append(mCoprs, dep.Name)
			} else if strings.HasPrefix(dep.FullName, "cm:") {
				mCms = append(mCms, dep.Name)
			}
		}

		for _, copr := range mCoprs {
			err := d.InstallCopr(ctx, copr)
			if err != nil {
				shared.PtermRemoved.Println(fmt.Sprintf(shared.PtermSpinnerStatusMsgs[shared.PtermSpinnerInstall].Fail, copr))
				// return nil, err
			} else {
				shared.PtermInstalled.Println(fmt.Sprintf(shared.PtermSpinnerStatusMsgs[shared.PtermSpinnerInstall].Success, copr))
			}
		}
		for _, cm := range mCms {
			err = shared.PtermSpinner(shared.PtermSpinnerInstall, cm, func() error {
				return d.InstallCm(ctx, cm)
			})
			if err != nil {
				log.Err(err).Str("manager", string(Name)).Str("dependency", cm)
				err = nil
			}
		}
	}

	fmt.Println("")
	for _, dep := range depStatus.Removed {
		if strings.HasPrefix(dep.FullName, "copr:") {
			err = shared.PtermSpinner(shared.PtermSpinnerRemove, dep.Name, func() error {
				return d.RemoveCopr(ctx, dep.Name)
			})
			if err != nil {
				log.Err(err).Str("manager", string(Name)).Str("dependency", dep.Name)
				err = nil
			}
		} else if strings.HasPrefix(dep.FullName, "cm:") {
			err = shared.PtermSpinner(shared.PtermSpinnerRemove, dep.Name, func() error {
				return d.RemoveCm(ctx, dep.Name)
			})
			if err != nil {
				log.Err(err).Str("manager", string(Name)).Str("dependency", dep.Name)
				err = nil
			}
		}
	}

	return
}

func (d *Dnf) SyncPackages(ctx context.Context, packageStatus status.PackageStatus) (userWarnings []string, err error) {
	if len(packageStatus.Missing) > 0 {
		filteredPkgsInstall := lo.Filter(packageStatus.Missing, func(item shared.Package, _ int) bool {
			isSysPkg, err := d.isSystemPackage(ctx, item.FullName)
			if err != nil || isSysPkg {
				return false
			}
			return true
		})

		fmt.Println("")
		err := d.InstallPkg(ctx, filteredPkgsInstall)
		if err != nil {
			return nil, err
		}
	}

	if len(packageStatus.Removed) > 0 {
		filteredPkgsRemove := lo.Filter(packageStatus.Removed, func(item shared.Package, _ int) bool {
			isSysPkg, err := d.isSystemPackage(ctx, item.FullName)
			if err != nil || isSysPkg {
				return false
			}
			return true
		})

		fmt.Println("")
		err := d.RemovePkg(ctx, filteredPkgsRemove)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (d *Dnf) isSystemPackage(ctx context.Context, pkg string) (bool, error) {
	allPkgs, _, err := d.ListInstalledPkgs(ctx)
	if err != nil {
		return false, err
	}

	userPkgs, err := d.ListUserInstalledPkgs(ctx)
	if err != nil {
		return false, err
	}

	if lo.Contains(allPkgs, pkg) && !lo.Contains(userPkgs, pkg) {
		return true, nil
	}

	return false, nil
}

func (d *Dnf) sortDeps(deps []string) (pCoprs, pCms []shared.Dependency) {
	for _, dep := range deps {
		sDep := strings.SplitN(dep, ":", 2)
		switch sDep[0] {
		case "copr":
			pCoprs = append(pCoprs, shared.Dependency{Name: sDep[1], FullName: dep})
		case "cm":
			pCms = append(pCms, shared.Dependency{Name: sDep[1], FullName: dep})
		default:
			shared.PtermWarning.Printfln("Dependency has bad format: %s. Ignoring...", dep)
		}
	}
	return
}

func (d *Dnf) coprFilename(coprName string) string {
	if strings.Count(coprName, "/") == 1 {
		return path.Join("copr.fedorainfracloud.org", coprName)
	}
	return coprName
}

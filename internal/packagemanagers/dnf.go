package packagemanagers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/alexellis/go-execute/v2"
	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Dnf struct {
	cacheAllInstalled  []string
	cacheUserInstalled []string
	cacheCoprs         []string
}

func (d *Dnf) Name() string {
	return "dnf"
}

func (d *Dnf) Icon() string {
	return ""
}

func (d *Dnf) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{shared.CommandInstall, shared.CommandRemove, shared.CommandSync}
}

func (d *Dnf) GetPackageNames(ctx context.Context, packagesConfig shared.PmPackages) []string {
	return packagesConfig.Global.Packages
}

func (d *Dnf) Add(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packageConfig shared.PmPackages, userWarnings []string, err error) {
	for _, pkg := range pkgs {
		isSysPkg, err := d.isSystemPackage(ctx, pkg)
		if err != nil {
			return shared.PmPackages{}, []string{}, err
		}

		if !isSysPkg {
			packagesConfig.Global.Packages = append(packagesConfig.Global.Packages, pkg)
		} else {
			userWarnings = append(userWarnings, fmt.Sprintf("'%s' is a system package and cannot be managed", pkg))
		}
	}
	return packagesConfig, userWarnings, nil
}

func (d *Dnf) InstallValidArgs(ctx context.Context, toComplete string) ([]string, error) {
	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"list", "--available", toComplete + "*"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, errors.New("Non-zero exit code: " + res.Stderr)
	}

	dnfList := strings.Split(strings.TrimSpace(res.Stdout), "\n")
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

func (d *Dnf) ListDependencies(ctx context.Context, tx *gorm.DB, packages shared.PmPackages) (depStatus shared.DependenciesStatus, err error) {
	installedCoprs, err := d.listCoprs(ctx)
	if err != nil {
		return shared.DependenciesStatus{}, err
	}

	installedCms, err := d.listCm(ctx)
	if err != nil {
		return shared.DependenciesStatus{}, err
	}

	pCoprs, pCms := d.sortDeps(packages.Global.Dependencies)

	// COPR
	for _, dep := range pCoprs {
		depFound := false
		for _, dnfDep := range installedCoprs {
			if dnfDep == dep.Name {
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

	stateDeps, err := state.GetDependencyState(tx, d.Name())
	if err != nil {
		return
	}

	sCoprs, sCms := d.sortDeps(stateDeps)

	// COPR
	for _, dep := range sCoprs {
		for _, dnfDep := range installedCoprs {
			if dnfDep == dep.Name {
				if !lo.Contains(packages.Global.Dependencies, dep.FullName) {
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
				if !lo.Contains(packages.Global.Dependencies, dep.FullName) {
					depStatus.Removed = append(depStatus.Removed, dep)
				}
				break
			}
		}
	}

	return
}

func (d *Dnf) ListPackages(ctx context.Context, tx *gorm.DB, packages shared.PmPackages) (packageStatus shared.PackageStatus, err error) {
	dnfList, err := d.listInstalled(ctx)
	if err != nil {
		return
	}

	for _, pkg := range packages.Global.Packages {
		pkgFound := false
		for _, dnfPkg := range dnfList {
			if dnfPkg == pkg {
				packageStatus.Synced = append(packageStatus.Synced, shared.Package{Name: pkg, FullName: pkg})
				// installedPkgs = append(installedPkgs, pkg)
				pkgFound = true
				break
			}
		}
		if !pkgFound {
			packageStatus.Missing = append(packageStatus.Missing, shared.Package{Name: pkg, FullName: pkg})
			// missingPkgs = append(missingPkgs, pkg)
		}
	}

	// FIXME: State check should be global for all managers
	// So removedPkgs should not be coming from this func
	//
	// NO! Scratch that. Ofc the manager needs to deal with it..
	// Otherwise we cant make sure the package is installed or not
	statePkgs, err := state.GetPackageState(tx, d.Name())
	if err != nil {
		return
		// return nil, nil, nil, err
	}

	for _, pkg := range statePkgs {
		for _, dnfPkg := range dnfList {
			if dnfPkg == pkg {
				if !lo.Contains(packages.Global.Packages, pkg) {
					packageStatus.Removed = append(packageStatus.Removed, shared.Package{Name: pkg, FullName: pkg})
					// removedPkgs = append(removedPkgs, pkg)
				}
				break
			}
		}
	}

	return
}

func (d *Dnf) Remove(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packageConfig shared.PmPackages, userWarnings []string, err error) {
	for _, pkg := range pkgs {
		isSysPkg, err := d.isSystemPackage(ctx, pkg)
		if err != nil {
			return shared.PmPackages{}, []string{}, err
		}

		if isSysPkg {
			userWarnings = append(userWarnings, fmt.Sprintf("'%s' is a system package and cannot be managed", pkg))
		}
	}

	packagesConfig.Global.Packages = lo.Filter(packagesConfig.Global.Packages, func(item string, index int) bool {
		return !lo.Contains(pkgs, item)
	})

	return packagesConfig, userWarnings, nil
}

func (d *Dnf) SyncDependencies(ctx context.Context, depStatus shared.DependenciesStatus) (userWarnings []string, err error) {
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
			err := d.installCopr(ctx, copr)
			if err != nil {
				shared.PtermRemoved.Println(fmt.Sprintf(shared.PtermSpinnerStatusMsgs[shared.PtermSpinnerInstall].Fail, copr))
				// return nil, err
			} else {
				shared.PtermInstalled.Println(fmt.Sprintf(shared.PtermSpinnerStatusMsgs[shared.PtermSpinnerInstall].Success, copr))
			}
		}
		for _, cm := range mCms {
			err = shared.PtermSpinner(shared.PtermSpinnerInstall, cm, func() error {
				return d.installCm(ctx, cm)
			})
			//NOTE: Not sure what to do with err here. Maybe just verbose log?
			err = nil
		}
	}

	fmt.Println("")
	for _, dep := range depStatus.Removed {
		if strings.HasPrefix(dep.FullName, "copr:") {
			err = shared.PtermSpinner(shared.PtermSpinnerRemove, dep.Name, func() error {
				return d.removeCopr(ctx, dep.Name)
			})
			//NOTE: Not sure what to do with err here. Maybe just verbose log?
			err = nil
		} else if strings.HasPrefix(dep.FullName, "cm:") {
			err = shared.PtermSpinner(shared.PtermSpinnerRemove, dep.Name, func() error {
				return d.removeCm(ctx, dep.Name)
			})
			//NOTE: Not sure what to do with err here. Maybe just verbose log?
			err = nil
		}
	}

	return
}

func (d *Dnf) SyncPackages(ctx context.Context, packageStatus shared.PackageStatus) (userWarnings []string, err error) {
	if len(packageStatus.Missing) > 0 {
		filteredPkgsInstall := lo.Filter(packageStatus.Missing, func(item shared.Package, _ int) bool {
			isSysPkg, err := d.isSystemPackage(ctx, item.FullName)
			if err != nil || isSysPkg {
				return false
			}
			return true
		})

		fmt.Println("")
		err := d.install(ctx, filteredPkgsInstall)
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
		err := d.remove(ctx, filteredPkgsRemove)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (d *Dnf) install(ctx context.Context, pkgs []shared.Package) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}

	pkgNames := []string{}
	for _, pkg := range pkgs {
		pkgNames = append(pkgNames, pkg.FullName)
	}

	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "install"}, pkgNames...),
		StreamStdio: true,
		Stdin:       os.Stdin,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return errors.New("Non-zero exit code: " + res.Stderr)
	}

	return nil
}

func (d *Dnf) remove(ctx context.Context, pkgs []shared.Package) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}

	pkgNames := []string{}
	for _, pkg := range pkgs {
		pkgNames = append(pkgNames, pkg.FullName)
	}

	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "remove"}, pkgNames...),
		StreamStdio: true,
		Stdin:       os.Stdin,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return errors.New("Non-zero exit code: " + res.Stderr)
	}

	return nil
}

func (d *Dnf) installCm(ctx context.Context, cms string) error {
	u, err := url.ParseRequestURI(cms)
	if err != nil {
		return fmt.Errorf("not an url: %s, %s", cms, err)
	}

	repoFileName := path.Join(d.yumRepoFolder(), fmt.Sprintf("%s%s", d.repoFilePrefix(), path.Base(u.Path)))
	cacheRepoFileName := path.Join(config.CacheDir, fmt.Sprintf("%s%s", d.repoFilePrefix(), path.Base(u.Path)))

	res, err := http.Get(cms)
	if err != nil {
		return fmt.Errorf("error making http request: %s", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("client: could not read response body: %s", err)
	}

	err = os.WriteFile(cacheRepoFileName, resBody, 0644)
	if err != nil {
		return fmt.Errorf("yum repo: could not write file %s: %s", repoFileName, err)
	}

	_, err = shared.Command(ctx, "sudo", []string{"chown", "root:root", cacheRepoFileName}, false, nil)
	if err != nil {
		return fmt.Errorf("could not chown %s: %s", cacheRepoFileName, err)
	}

	_, err = shared.Command(ctx, "sudo", []string{"mv", cacheRepoFileName, repoFileName}, false, nil)
	if err != nil {
		return fmt.Errorf("could not move %s: %s", repoFileName, err)
	}

	return nil
}

func (d *Dnf) listCm(ctx context.Context) (packages []string, err error) {
	cms, err := os.ReadDir("/etc/yum.repos.d/")
	if err != nil {
		return []string{}, err
	}

	for _, e := range cms {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), d.repoFilePrefix()) {
			packages = append(packages, strings.ReplaceAll(e.Name(), d.repoFilePrefix(), ""))
		}
	}
	return
}

func (d *Dnf) removeCm(ctx context.Context, cm string) error {
	u, err := url.ParseRequestURI(cm)
	if err != nil {
		return fmt.Errorf("not an url: %s, %s", cm, err)
	}

	repoFileName := path.Join(d.yumRepoFolder(), fmt.Sprintf("%s%s", d.repoFilePrefix(), path.Base(u.Path)))

	_, err = os.Stat(repoFileName)
	if os.IsNotExist(err) {
		return fmt.Errorf("remove cm: %s, file does not exist", cm)
	}

	_, err = shared.Command(ctx, "sudo", []string{"rm", repoFileName}, false, nil)
	return err
}

func (d *Dnf) installCopr(ctx context.Context, copr string) error {
	_, err := shared.Command(ctx, "sudo", append([]string{"dnf", "copr", "enable", copr}), true, os.Stdin)
	return err
}

func (d *Dnf) removeCopr(ctx context.Context, copr string) error {
	_, err := shared.Command(ctx, "sudo", []string{"dnf", "copr", "remove", copr}, false, nil)
	return err
}

func (d *Dnf) listCoprs(ctx context.Context) ([]string, error) {
	if len(d.cacheCoprs) > 0 {
		return d.cacheCoprs, nil
	}

	ret, err := shared.Command(ctx, "dnf", []string{"copr", "list"}, false, nil)
	if err != nil {
		return nil, err
	}

	lo.ForEach(strings.Split(strings.TrimSpace(ret), "\n"), func(item string, _ int) {
		d.cacheCoprs = append(d.cacheCoprs, strings.TrimSpace(item))
	})

	return d.cacheCoprs, nil
}

func (d *Dnf) listInstalled(ctx context.Context) ([]string, error) {
	if len(d.cacheAllInstalled) > 0 {
		return d.cacheAllInstalled, nil
	}

	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"list", "installed"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, errors.New("Non-zero exit code: " + res.Stderr)
	}

	dnfList := strings.Split(res.Stdout, "\n")
	for _, pkg := range dnfList[1:] {
		d.cacheAllInstalled = append(d.cacheAllInstalled, strings.Split(pkg, ".")[0])
	}

	return d.cacheAllInstalled, nil
}

func (d *Dnf) listUserInstalled(ctx context.Context) ([]string, error) {
	if len(d.cacheUserInstalled) > 0 {
		return d.cacheUserInstalled, nil
	}

	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"repoquery", "--userinstalled", "--qf", "%{name} %{version}"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, errors.New("Non-zero exit code: " + res.Stderr)
	}

	dnfList := strings.Split(res.Stdout, "\n")
	for _, pkg := range dnfList {
		d.cacheUserInstalled = append(d.cacheUserInstalled, strings.Split(pkg, " ")[0])
	}

	return d.cacheUserInstalled, nil
}

func (d *Dnf) isSystemPackage(ctx context.Context, pkg string) (bool, error) {
	allPkgs, err := d.listInstalled(ctx)
	if err != nil {
		return false, err
	}

	userPkgs, err := d.listUserInstalled(ctx)
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

func (d *Dnf) yumRepoFolder() string {
	return "/etc/yum.repos.d"
}

func (d *Dnf) repoFilePrefix() string {
	return "_packtrak:"
}

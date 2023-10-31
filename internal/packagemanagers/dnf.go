package packagemanagers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alexellis/go-execute/v2"
	"github.com/lucas-ingemar/mdnf/internal/shared"
	"github.com/samber/lo"
)

// FIXME: ska formodligen spara ner installed, missing och removed i strucet for att kunna
// accessa senare. Blir battre i printen
type Dnf struct {
	Lucas              string
	Banan              string
	cacheAllInstalled  []string
	cacheUserInstalled []string
}

func (d *Dnf) Name() string {
	// return "dnf"
	return d.Lucas
}

func (d *Dnf) Icon() string {
	// return "󰟓"
	// return "󰊢"
	// return ""
	return d.Banan
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

func (d *Dnf) List(ctx context.Context, packages shared.PmPackages, state shared.State) (installedPkgs []string, missingPkgs []string, removedPkgs []string, err error) {
	dnfList, err := d.listInstalled(ctx)
	if err != nil {
		return
	}

	for _, pkg := range packages.Global.Packages {
		pkgFound := false
		for _, dnfPkg := range dnfList {
			if dnfPkg == pkg {
				installedPkgs = append(installedPkgs, pkg)
				pkgFound = true
				break
			}
		}
		if !pkgFound {
			missingPkgs = append(missingPkgs, pkg)
		}
	}

	// FIXME: State check should be global for all managers
	// So removedPkgs should not be coming from this func
	//
	// NO! Scratch that. Ofc the manager needs to deal with it..
	// Otherwise we cant make sure the package is installed or not
	for _, pkg := range state.Packages[d.Name()].Global.Packages {
		for _, dnfPkg := range dnfList {
			if dnfPkg == pkg {
				if !lo.Contains(packages.Global.Packages, pkg) {
					removedPkgs = append(removedPkgs, pkg)
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

func (d *Dnf) Sync(ctx context.Context, pkgsInstall, pkgsRemove []string) (userWarnings []string, err error) {
	if len(pkgsInstall) > 0 {
		filteredPkgsInstall := lo.Filter(pkgsInstall, func(item string, _ int) bool {
			isSysPkg, err := d.isSystemPackage(ctx, item)
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

	if len(pkgsRemove) > 0 {
		filteredPkgsRemove := lo.Filter(pkgsRemove, func(item string, _ int) bool {
			isSysPkg, err := d.isSystemPackage(ctx, item)
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

func (d *Dnf) install(ctx context.Context, pkgs []string) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}
	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "install"}, pkgs...),
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

func (d *Dnf) remove(ctx context.Context, pkgs []string) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}
	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "remove"}, pkgs...),
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

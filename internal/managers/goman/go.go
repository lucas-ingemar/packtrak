package goman

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/samber/lo"
)

func New() *Go {
	return &Go{
		&commandExecutor{},
	}
}

const Name shared.ManagerName = "go"

type Go struct {
	CommandExecutorFace
}

func (g *Go) Name() shared.ManagerName {
	return Name
}

func (g *Go) Icon() string {
	return "󰟓"
}

func (g *Go) ShortDesc() string {
	return "Compile and install Go packages"
}

func (g *Go) LongDesc() string {
	return "Compile and install remote Go packages. It's added to the path and ready for you to use"
}

func (g *Go) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{}
}

func (g *Go) InitCheckCmd() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return errors.New("'go' command not found on the computer")
	}
	vString, err := shared.Command(context.Background(), "go", []string{"version"}, false, nil)
	if err != nil {
		return errors.New("failed to execute 'go version'")
	}

	r, err := regexp.Compile(`go(\d+\.\d+\.\d+)`)
	if err != nil {
		return errors.New("failed to match version")
	}
	v := r.FindStringSubmatch(vString)[1]
	if v < "1.18.0" {
		return fmt.Errorf("must use a go v1.18.0 or later. Current version %s", v)
	}
	return nil
}

func (g *Go) GetPackageNames(ctx context.Context, packages []string) []string {
	pkgNames := []string{}
	for _, pkg := range packages {
		pkgNames = append(pkgNames, g.nameFromFullName(pkg))
	}
	return pkgNames
}

func (g *Go) GetDependencyNames(ctx context.Context, deps []string) []string {
	return []string{}
}

func (g *Go) AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error) {
	// FIXME: Could do something more fancy perhaps? See if the path exists and so on
	// packagesConfig.Global.Packages = append(packagesConfig.Global.Packages, pkgs...)
	return pkgsToAdd, nil, nil
}

func (g *Go) AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (g *Go) InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error) {
	return []string{}, nil
}

func (g *Go) ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus status.DependenciesStatus, err error) {
	return
}

func (g *Go) ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus status.PackageStatus, err error) {
	installed, err := g.ListInstalled(ctx)
	if err != nil {
		return
	}

	for _, pkgFullName := range packages {
		pkgName := g.nameFromFullName(pkgFullName)
		iPkg, err := shared.GetPackage(pkgName, installed)
		if err != nil {
			packageStatus.Missing = append(packageStatus.Missing, shared.Package{
				Name:     pkgName,
				FullName: pkgFullName,
			})
			continue
		}
		dPkg, err := shared.GetDepsDevDefaultPackage(string(g.Name()), iPkg)
		if err != nil {
			return packageStatus, err
		}

		if iPkg.Version == dPkg.Version {
			packageStatus.Synced = append(packageStatus.Synced, iPkg)
		} else {
			packageStatus.Updated = append(packageStatus.Updated, shared.Package{
				Name:          iPkg.Name,
				FullName:      iPkg.FullName,
				Version:       iPkg.Version,
				LatestVersion: dPkg.Version,
				RepoUrl:       iPkg.RepoUrl,
			})
		}
	}

	for _, pkg := range statePkgs {
		if !lo.Contains(packages, pkg) {
			packageStatus.Removed = append(packageStatus.Removed, shared.Package{
				Name:     g.nameFromFullName(pkg),
				FullName: pkg,
			})
		}
	}

	return
}

func (g *Go) RemovePackages(ctx context.Context, allPkgs []string, pkgs []string) (packagesUpdated []string, userWarnings []string, err error) {
	binPath, err := g.BinPath()
	if err != nil {
		return nil, nil, err
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg)
		pkgObj, err := g.GetBinaryInfo(ctx, path.Join(binPath, pkg))
		if err != nil {
			return nil, nil, err
		}
		packagesUpdated = append(packagesUpdated, pkgObj.FullName)
	}
	return
}

func (g *Go) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (g *Go) SyncDependencies(ctx context.Context, depStatus status.DependenciesStatus) (userWarnings []string, err error) {
	return
}

func (g *Go) SyncPackages(ctx context.Context, packageStatus status.PackageStatus) (userWarnings []string, err error) {
	for _, pkg := range packageStatus.Missing {
		err = shared.PtermSpinner(shared.PtermSpinnerInstall, pkg.Name, func() error {
			return g.Install(ctx, pkg)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}

	for _, pkg := range packageStatus.Updated {
		err = shared.PtermSpinner(shared.PtermSpinnerUpdate, pkg.Name, func() error {
			return g.Install(ctx, pkg)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}

	for _, pkg := range packageStatus.Removed {
		err = shared.PtermSpinner(shared.PtermSpinnerRemove, pkg.Name, func() error {
			return g.Remove(pkg)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}
	return
}

func (g *Go) nameFromFullName(fullName string) string {
	cmps := strings.Split(fullName, "/")
	matched, err := regexp.MatchString(`^v(\d+\.)?(\d+\.)?(\*|\d+)$`, cmps[len(cmps)-1])
	if err != nil {
		return fullName
	}

	if matched {
		return cmps[len(cmps)-2]
	}
	return cmps[len(cmps)-1]
}

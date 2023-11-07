package packagemanagers

import (
	"context"
	"errors"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/alexellis/go-execute/v2"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

const (
	goVersionCheckBaseUrl = "https://api.deps.dev/v3alpha/systems/go/packages"
)

type Go struct {
}

func (g *Go) Name() string {
	return "go"
}

func (g *Go) Icon() string {
	return "󰟓"
}

func (g *Go) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{}
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

func (g *Go) InstallValidArgs(ctx context.Context, toComplete string) ([]string, error) {
	return []string{}, nil
}

func (g *Go) ListDependencies(ctx context.Context, tx *gorm.DB, deps []string) (depStatus shared.DependenciesStatus, err error) {
	return
}

func (g *Go) ListPackages(ctx context.Context, tx *gorm.DB, packages []string) (packageStatus shared.PackageStatus, err error) {
	installed, err := g.listInstalled(ctx)
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
		dPkg, err := shared.GetDepsDevDefaultPackage(g.Name(), iPkg)
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

	statePkgs, err := state.GetPackageState(tx, "go")
	if err != nil {
		return shared.PackageStatus{}, err
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
	for _, pkg := range pkgs {
		packagesUpdated = append(packagesUpdated, g.nameFromFullName(pkg))
	}
	// packagesToRemove = lo.Filter(packages, func(item string, index int) bool {
	// 	return lo.Contains(pkgs, g.nameFromFullName(item))
	// })
	return
}

func (g *Go) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (g *Go) SyncDependencies(ctx context.Context, depStatus shared.DependenciesStatus) (userWarnings []string, err error) {
	return
}

func (g *Go) SyncPackages(ctx context.Context, packageStatus shared.PackageStatus) (userWarnings []string, err error) {
	// fmt.Println(packageStatus)
	for _, pkg := range packageStatus.Missing {
		err = shared.PtermSpinner(shared.PtermSpinnerInstall, pkg.Name, func() error {
			return g.install(ctx, pkg)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}

	for _, pkg := range packageStatus.Updated {
		err = shared.PtermSpinner(shared.PtermSpinnerUpdate, pkg.Name, func() error {
			return g.install(ctx, pkg)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}

	for _, pkg := range packageStatus.Removed {
		err = shared.PtermSpinner(shared.PtermSpinnerRemove, pkg.Name, func() error {
			return g.remove(pkg)
		})
		//NOTE: Not sure what to do with err here. Maybe just verbose log?
		err = nil
	}
	return
}

func (g *Go) install(ctx context.Context, pkg shared.Package) error {
	_, err := shared.Command(ctx, "go", []string{"install", pkg.FullName + "@latest"}, false, nil)
	if err != nil {
		return err
	}
	return nil
}

func (g *Go) remove(pkg shared.Package) error {
	binPath, err := g.binPath()
	if err != nil {
		return err
	}

	pkgPath := path.Join(binPath, pkg.Name)

	binary, err := os.Stat(pkgPath)
	if err != nil {
		return err
	}

	if binary.IsDir() {
		return errors.New("not a file")
	}

	return os.Remove(pkgPath)
}

func (g *Go) listInstalled(ctx context.Context) (packages []shared.Package, err error) {
	binPath, err := g.binPath()
	if err != nil {
		return nil, err
	}
	binaries, err := os.ReadDir(binPath)
	if err != nil {
		return
	}

	for _, e := range binaries {
		if e.IsDir() {
			continue
		}

		pkg, err := g.getBinaryInfo(ctx, path.Join(binPath, e.Name()))
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return
}

func (g *Go) binPath() (binPath string, err error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return "", errors.New("GOPATH not found")
	}
	return path.Join(goPath, "bin"), nil
}

func (g *Go) getBinaryInfo(ctx context.Context, binaryPath string) (pkg shared.Package, err error) {
	cmd := execute.ExecTask{
		Command:     "go",
		Args:        []string{"version", "-m", binaryPath},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return
	}

	if res.ExitCode != 0 {
		return pkg, errors.New("Non-zero exit code: " + res.Stderr)
	}

	rPath, err := regexp.Compile(`(?m)^\s*path\s*(\S+)$`)
	if err != nil {
		return
	}

	rVersion, err := regexp.Compile(`(?m)^\s*mod\s*(\S+)\s*(\S+)\s*\S+$`)
	if err != nil {
		return
	}

	pathMatches := rPath.FindStringSubmatch(res.Stdout)
	if len(pathMatches) != 2 {
		return pkg, errors.New("could not match path")
	}

	versionMatches := rVersion.FindStringSubmatch(res.Stdout)
	if len(versionMatches) != 3 {
		return pkg, errors.New("could not match version")
	}

	_, name := path.Split(binaryPath)

	return shared.Package{
		Name:     name,
		FullName: pathMatches[1],
		Version:  versionMatches[2],
		RepoUrl:  versionMatches[1],
	}, nil
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

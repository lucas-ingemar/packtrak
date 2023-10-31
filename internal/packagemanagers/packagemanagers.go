package packagemanagers

import (
	"context"

	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/shared"
)

var (
	PackageManagers = []PackageManager{&Dnf{Lucas: "dnf", Banan: ""}, &Dnf{Lucas: "git", Banan: "󰊢"}}
)

func MustInitPackages() shared.Packages {
	var err error
	config.Packages, err = config.ReadPackagesConfig()
	if err != nil {
		panic(err)
	}

	for _, pm := range PackageManagers {
		_, ok := config.Packages[pm.Name()]
		if !ok {
			config.Packages[pm.Name()] = shared.PmPackages{}
		}
	}
	return config.Packages
}

type PackageManager interface {
	Name() string
	Icon() string

	Add(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packageConfig shared.PmPackages, userWarnings []string, err error)
	InstallValidArgs(ctx context.Context, toComplete string) ([]string, error)
	List(ctx context.Context, packages shared.PmPackages, state shared.State) (installedPkgs []string, missingPkgs []string, removedPkgs []string, err error)
	Remove(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packageConfig shared.PmPackages, userWarnings []string, err error)
	Sync(ctx context.Context, pkgsInstall, pkgsRemove []string) (userWarnings []string, err error)
}

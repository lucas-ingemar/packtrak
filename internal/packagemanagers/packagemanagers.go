package packagemanagers

import (
	"context"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"gorm.io/gorm"
)

var (
	PackageManagersRegistered= []PackageManager{&Dnf{Lucas: "dnf", Banan: ""}, &Dnf{Lucas: "git", Banan: "󰊢"}}
	PackageManagers = []PackageManager{}
)

func InitPackageManagers() {
	for _, pm := range PackageManagersRegistered {
		//FIXME: HERE should the enabled/disabled flag be
		PackageManagers = append(PackageManagers, pm)
	}
}

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
	List(ctx context.Context, tx *gorm.DB, packages shared.PmPackages) (installedPkgs []string, missingPkgs []string, removedPkgs []string, err error)
	Remove(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packageConfig shared.PmPackages, userWarnings []string, err error)
	Sync(ctx context.Context, pkgsInstall, pkgsRemove []string) (userWarnings []string, err error)
}

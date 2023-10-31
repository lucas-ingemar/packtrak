package packagemanagers

import (
	"context"

	"github.com/lucas-ingemar/mdnf/internal/shared"
)

var (
	PackageManagers = []PackageManager{&Dnf{Lucas: "dnf"}, &Dnf{Lucas: "git"}}
)

type PackageManager interface {
	Name() string
	Icon() string

	Add(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packageConfig shared.PmPackages, userWarnings []string, err error)
	InstallValidArgs(ctx context.Context, toComplete string) ([]string, error)
	List(ctx context.Context, packages shared.PmPackages, state shared.State) (installedPkgs []string, missingPkgs []string, removedPkgs []string, err error)
	Remove(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packageConfig shared.PmPackages, userWarnings []string, err error)
	Sync(ctx context.Context, pkgsInstall, pkgsRemove []string) (userWarnings []string, err error)
}

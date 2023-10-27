package mdnf

import (
	"github.com/lucas-ingemar/mdnf/internal/shared"
	"github.com/samber/lo"
)

func AddPackages(packagesConfig shared.Packages, pkgs []string) (shared.Packages, error) {
	packagesConfig.Dnf.Global.Packages = append(packagesConfig.Dnf.Global.Packages, pkgs...)
	return packagesConfig, nil
}

func RemovePackages(packagesConfig shared.Packages, pkgs []string) (shared.Packages, error) {
	packagesConfig.Dnf.Global.Packages = lo.Filter(packagesConfig.Dnf.Global.Packages, func(item string, index int) bool {
		return !lo.Contains(pkgs, item)
	})
	return packagesConfig, nil
}
